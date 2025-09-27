package plugin

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	reduct "github.com/reductstore/reduct-go"
	"github.com/stretchr/testify/assert"
)

func getServerUrl() string {
	url := os.Getenv("REDUCT_URL")
	if url == "" {
		url = "http://127.0.0.1:8383"
	}
	return url
}

func runQuery(tb testing.TB, query string) (backend.QueryDataResponse, func(tb testing.TB)) {
	// Prepare bucket with data
	ctx := context.Background()

	client := reduct.NewClient(getServerUrl(), reduct.ClientOptions{})
	bucket, err := client.CreateOrGetBucket(ctx, "test-bucket", nil)
	if err != nil {
		tb.Fatal(err)
	}

	ts := time.Now().UnixMicro()

	for i := int64(0); i < 10; i++ {
		record := bucket.BeginWrite(ctx, "entity1", &reduct.WriteOptions{
			Timestamp: ts + i,
			Labels: map[string]any{
				"bool-label":   bool(i%2 == 0),
				"int-label":    i,
				"float-label":  float64(i) + 0.5,
				"string-label": "label-" + strconv.FormatInt(i, 10),
			},
		})

		err := record.Write("any")

		if err != nil {
			tb.Fatal(err)
		}

	}

	instance, err := NewDatasource(ctx, backend.DataSourceInstanceSettings{
		JSONData: json.RawMessage(`{
			"serverURL": "` + getServerUrl() + `",
			"verifySSL": false
		}`),
	})

	assert.Nil(tb, err)

	ds := instance.(*ReductDatasource)

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{
					RefID: "A",
					TimeRange: backend.TimeRange{
						From: time.UnixMicro(ts).Add(-time.Minute),
						To:   time.UnixMicro(ts + 10).Add(time.Minute),
					},
					JSON: json.RawMessage(query),
				},
			},
		},
	)
	if err != nil {
		tb.Error(err)
	}

	return *resp, func(tb testing.TB) {
		// Cleanup bucket after test
		err := bucket.Remove(ctx)
		if err != nil {
			tb.Fatal(err)
		}
	}
}

func TestQueryData(t *testing.T) {
	resp, teardown := runQuery(t, `{
						"Bucket": "test-bucket",
						"Entry": "entity1"
					}`)
	defer teardown(t)

	assert.Equal(t, nil, resp.Responses["A"].Error)

	assert.Equal(t, 4, len(resp.Responses["A"].Frames))

	idx := findByName(&resp, "bool-label")
	assert.Equal(t, 10, resp.Responses["A"].Frames[idx].Rows())
	assert.Equal(t, "bool-label", resp.Responses["A"].Frames[idx].Name)
	assert.Equal(t, true, resp.Responses["A"].Frames[idx].Fields[1].At(0))
	assert.Equal(t, false, resp.Responses["A"].Frames[idx].Fields[1].At(9))

	idx = findByName(&resp, "int-label")
	assert.Equal(t, 10, resp.Responses["A"].Frames[idx].Rows())
	assert.Equal(t, "int-label", resp.Responses["A"].Frames[idx].Name)
	assert.Equal(t, int64(0), resp.Responses["A"].Frames[idx].Fields[1].At(0))
	assert.Equal(t, int64(9), resp.Responses["A"].Frames[idx].Fields[1].At(9))

	idx = findByName(&resp, "float-label")
	assert.Equal(t, 10, resp.Responses["A"].Frames[idx].Rows())
	assert.Equal(t, "float-label", resp.Responses["A"].Frames[idx].Name)
	assert.Equal(t, 0.5, resp.Responses["A"].Frames[idx].Fields[1].At(0))
	assert.Equal(t, 9.5, resp.Responses["A"].Frames[idx].Fields[1].At(9))

	idx = findByName(&resp, "string-label")
	assert.Equal(t, 10, resp.Responses["A"].Frames[idx].Rows())
	assert.Equal(t, "string-label", resp.Responses["A"].Frames[idx].Name)
	assert.Equal(t, "label-0", resp.Responses["A"].Frames[idx].Fields[1].At(0))
	assert.Equal(t, "label-9", resp.Responses["A"].Frames[idx].Fields[1].At(9))
}

func TestQueryDataWithThen(t *testing.T) {
	resp, teardown := runQuery(t, `{
						"Bucket": "test-bucket",
						"Entry": "entity1",
						"Options": {
							"When": { "#select_labels": ["int-label"]},
							"Mode": "labels"
						}
					}`)
	defer teardown(t)

	assert.Equal(t, nil, resp.Responses["A"].Error)
	assert.Equal(t, 1, len(resp.Responses["A"].Frames))

	assert.Equal(t, "int-label", resp.Responses["A"].Frames[0].Name)
	assert.Equal(t, 10, resp.Responses["A"].Frames[0].Rows())
}

func TestQueryDataBadFormat(t *testing.T) {
	resp, teardown := runQuery(t, `{broken}`)
	defer teardown(t)

	assert.Equal(t, backend.ErrDataResponse(backend.StatusBadRequest, "invalid query format"), resp.Responses["A"])
}

func TestQueryDataNoBucket(t *testing.T) {
	resp, teardown := runQuery(t, `{
						"Entry": "entity1"
					}`)
	defer teardown(t)
	assert.Equal(t, backend.ErrDataResponse(backend.StatusBadRequest, "missing bucket or entry"), resp.Responses["A"])
}

func TestQueryDataNoEntry(t *testing.T) {
	resp, teardown := runQuery(t, `{
						"Bucket": "test-bucket"
					}`)
	defer teardown(t)
	assert.Equal(t, backend.ErrDataResponse(backend.StatusBadRequest, "missing bucket or entry"), resp.Responses["A"])
}

func TestQueryDataBucketNotFound(t *testing.T) {
	resp, teardown := runQuery(t, `{
						"Bucket": "missing-bucket",
						"Entry": "entity1"
					}`)
	defer teardown(t)
	assert.Equal(t, backend.ErrDataResponse(backend.StatusNotFound, "bucket 'missing-bucket' not found"), resp.Responses["A"])
}

func TestQueryDataEntryNotFound(t *testing.T) {
	resp, teardown := runQuery(t, `{
						"Bucket": "test-bucket",
						"Entry": "missing-entity"
					}`)
	defer teardown(t)
	assert.Equal(t, backend.ErrDataResponse(backend.StatusNotFound, "Entry 'missing-entity' not found in bucket 'test-bucket'"), resp.Responses["A"])

}

func findByName(resp *backend.QueryDataResponse, name string) int {
	idx := -1
	for i, frame := range resp.Responses["A"].Frames {
		if frame.Name == name {
			idx = i
		}
	}
	return idx
}
