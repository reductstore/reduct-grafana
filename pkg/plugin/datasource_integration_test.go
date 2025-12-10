//go:build integration
// +build integration

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
	ctx := context.Background()

	token := "dev-token"

	client := reduct.NewClient(getServerUrl(), reduct.ClientOptions{
		APIToken: token,
	})

	bucket, err := client.CreateOrGetBucket(ctx, "test-bucket", nil)
	if err != nil {
		tb.Fatal(err)
	}

	ts := time.Now().UnixMicro()

	for i := int64(0); i < 10; i++ {
		record := bucket.BeginWrite(ctx, "entity1", &reduct.WriteOptions{
			Timestamp: ts + i,
			Labels: map[string]any{
				"bool-label":   i%2 == 0,
				"int-label":    i,
				"float-label":  float64(i) + 0.5,
				"string-label": "label-" + strconv.FormatInt(i, 10),
			},
		})

		b, _ := json.Marshal(map[string]any{
			"temp":       float64(i) + 0.25,
			"flag":       i%2 == 0,
			"meta":       map[string]any{"seq": i},
			"str_number": "123",
			"source_id":  "00000001_000",
		})

		if err := record.Write(string(b)); err != nil {
			tb.Fatal(err)
		}
	}

	instance, err := NewDatasource(ctx, backend.DataSourceInstanceSettings{
		JSONData: json.RawMessage(`{
						"serverURL": "` + getServerUrl() + `",
						"verifySSL": false
				}`),
		DecryptedSecureJSONData: map[string]string{
			"serverToken": token,
		},
	})
	if err != nil {
		tb.Fatal(err)
	}

	ds := instance.(*ReductDatasource)

	resp, err := ds.QueryData(
		ctx,
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
		tb.Fatal(err)
	}

	return *resp, func(tb testing.TB) {
		if err := bucket.Remove(ctx); err != nil {
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
							"Mode": "LabelOnly"
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

func TestQueryData_ContentMode_ParsesJSON(t *testing.T) {
	resp, teardown := runQuery(t, `{
		"Bucket": "test-bucket",
		"Entry": "entity1",
		"Options": { "Mode": "ContentOnly" }
	}`)
	defer teardown(t)

	dr := resp.Responses["A"]
	assert.Nil(t, dr.Error)

	idx := findByName(&resp, "$.temp")
	if idx == -1 {
		t.Fatalf("frame $.temp not found")
	}
	assert.Equal(t, 10, dr.Frames[idx].Rows())
	assert.Equal(t, 0.25, dr.Frames[idx].Fields[1].At(0))
	assert.Equal(t, 9.25, dr.Frames[idx].Fields[1].At(9))

	idx = findByName(&resp, "$.flag")
	if idx == -1 {
		t.Fatalf("frame $.flag not found")
	}
	assert.Equal(t, true, dr.Frames[idx].Fields[1].At(0))
	assert.Equal(t, false, dr.Frames[idx].Fields[1].At(9))

	idx = findByName(&resp, "$.meta.seq")
	if idx == -1 {
		t.Fatalf("frame $.meta.seq not found")
	}
	assert.Equal(t, float64(0), dr.Frames[idx].Fields[1].At(0))
	assert.Equal(t, float64(9), dr.Frames[idx].Fields[1].At(9))
}

func TestQueryData_BothMode_LabelsAndJSON(t *testing.T) {
	resp, teardown := runQuery(t, `{
		"Bucket": "test-bucket",
		"Entry": "entity1",
		"Options": { "Mode": "LabelAndContent" }
	}`)
	defer teardown(t)

	dr := resp.Responses["A"]
	assert.Nil(t, dr.Error)

	idxLabel := findByName(&resp, "int-label")
	if idxLabel == -1 {
		t.Fatalf("label frame 'int-label' not found")
	}
	assert.Equal(t, 10, dr.Frames[idxLabel].Rows())
	assert.Equal(t, int64(0), dr.Frames[idxLabel].Fields[1].At(0))
	assert.Equal(t, int64(9), dr.Frames[idxLabel].Fields[1].At(9))

	idxJSON := findByName(&resp, "$.temp")
	if idxJSON == -1 {
		t.Fatalf("json frame '$.temp' not found")
	}
	assert.Equal(t, 0.25, dr.Frames[idxJSON].Fields[1].At(0))
	assert.Equal(t, 9.25, dr.Frames[idxJSON].Fields[1].At(9))
}

func TestQueryData_LabelsMode_IgnoresJSON(t *testing.T) {
	resp, teardown := runQuery(t, `{
		"Bucket": "test-bucket",
		"Entry": "entity1",
		"Options": { "Mode": "LabelOnly" }
	}`)
	defer teardown(t)

	dr := resp.Responses["A"]
	assert.Nil(t, dr.Error)

	assert.Equal(t, -1, findByName(&resp, "$.temp"))
	assert.Equal(t, -1, findByName(&resp, "$.flag"))
	assert.Equal(t, -1, findByName(&resp, "$.meta.seq"))

	idx := findByName(&resp, "string-label")
	if idx == -1 {
		t.Fatalf("label frame 'string-label' not found")
	}
	assert.Equal(t, 10, dr.Frames[idx].Rows())
	assert.Equal(t, "label-0", dr.Frames[idx].Fields[1].At(0))
	assert.Equal(t, "label-9", dr.Frames[idx].Fields[1].At(9))
}

func TestQueryData_ContentMode_PreservesJSONStringTypes(t *testing.T) {
	resp, teardown := runQuery(t, `{
		"Bucket": "test-bucket",
		"Entry": "entity1",
		"Options": { "Mode": "ContentOnly" }
	}`)
	defer teardown(t)

	dr := resp.Responses["A"]
	assert.Nil(t, dr.Error)

	idx := findByName(&resp, "$.str_number")
	if idx == -1 {
		t.Fatalf("frame $.str_number not found")
	}
	assert.Equal(t, 10, dr.Frames[idx].Rows())
	assert.Equal(t, "123", dr.Frames[idx].Fields[1].At(0))

	idx = findByName(&resp, "$.source_id")
	if idx == -1 {
		t.Fatalf("frame $.source_id not found")
	}
	assert.Equal(t, 10, dr.Frames[idx].Rows())
	assert.Equal(t, "00000001_000", dr.Frames[idx].Fields[1].At(0))
}
