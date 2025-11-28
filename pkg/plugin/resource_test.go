//go:build integration
// +build integration

package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/stretchr/testify/assert"
)

type testSender struct {
	resp *backend.CallResourceResponse
}

func (s *testSender) Send(r *backend.CallResourceResponse) error {
	*s.resp = *r
	return nil
}

func newTestDatasource(t *testing.T) *ReductDatasource {
	instance, err := NewDatasource(context.Background(), backend.DataSourceInstanceSettings{
		JSONData: json.RawMessage(`{
			"serverURL": "` + getServerUrl() + `",
			"verifySSL": false
		}`),
		DecryptedSecureJSONData: map[string]string{
			"serverToken": "dev-token",
		},
	})
	assert.NoError(t, err)
	return instance.(*ReductDatasource)
}

func newAdminClient() reductgo.Client {
	return reductgo.NewClient(
		getServerUrl(),
		reductgo.ClientOptions{
			APIToken: "dev-token",
		},
	)
}

func TestCallResource_ListBuckets(t *testing.T) {
	ds := newTestDatasource(t)

	client := newAdminClient()
	_, _ = client.CreateOrGetBucket(context.Background(), "cr-bucket", nil)

	var resp backend.CallResourceResponse
	sender := &testSender{resp: &resp}

	err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{
			Path: "listBuckets",
		},
		sender,
	)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Status)

	var buckets []map[string]any
	json.Unmarshal(resp.Body, &buckets)
	assert.NotEmpty(t, buckets)
}

func TestCallResource_ListEntries(t *testing.T) {
	ds := newTestDatasource(t)

	client := newAdminClient()
	bucket, _ := client.CreateOrGetBucket(context.Background(), "cr-entry-bucket", nil)
	bucket.BeginWrite(context.Background(), "entry1", nil).Write("x")

	body := []byte(`{"bucket": "cr-entry-bucket"}`)

	var resp backend.CallResourceResponse
	sender := &testSender{resp: &resp}

	err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{
			Path: "listEntries",
			Body: body,
		},
		sender,
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Status)

	var entries []map[string]any
	json.Unmarshal(resp.Body, &entries)
	assert.NotEmpty(t, entries)
}

func TestCallResource_ValidateCondition_Valid(t *testing.T) {
	ds := newTestDatasource(t)

	client := newAdminClient()
	bucket, _ := client.CreateOrGetBucket(context.Background(), "cr-val-bucket", nil)
	bucket.BeginWrite(context.Background(), "entity", nil).Write("123")

	body := []byte(`{
		"bucket": "cr-val-bucket",
		"entry": "entity",
		"condition": {"&sensor":{"$eq":"ok"}}
	}`)

	var resp backend.CallResourceResponse
	sender := &testSender{resp: &resp}

	err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{
			Path: "validateCondition",
			Body: body,
		},
		sender,
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Status)

	var parsed map[string]any
	json.Unmarshal(resp.Body, &parsed)

	assert.Equal(t, true, parsed["valid"])
}

func TestCallResource_ValidateCondition_InvalidJSON(t *testing.T) {
	ds := newTestDatasource(t)

	body := []byte(`{
		"bucket": "cr-val-bucket",
		"entry": "entity",
		"condition": "not-json"
	}`)

	var resp backend.CallResourceResponse
	sender := &testSender{resp: &resp}

	err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{
			Path: "validateCondition",
			Body: body,
		},
		sender,
	)
	assert.NoError(t, err)

	var parsed map[string]any
	json.Unmarshal(resp.Body, &parsed)

	assert.Equal(t, false, parsed["valid"])
}

func TestCallResource_ValidateCondition_MissingBucket(t *testing.T) {
	ds := newTestDatasource(t)

	body := []byte(`{"entry": "x", "condition": {}}`)

	var resp backend.CallResourceResponse
	sender := &testSender{resp: &resp}

	err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{
			Path: "validateCondition",
			Body: body,
		},
		sender,
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.Status)
}

func TestCallResource_ServerInfo(t *testing.T) {
	ds := newTestDatasource(t)

	var resp backend.CallResourceResponse
	sender := &testSender{resp: &resp}

	err := ds.CallResource(
		context.Background(),
		&backend.CallResourceRequest{
			Path: "serverInfo",
		},
		sender,
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.Status)

	var info map[string]any
	json.Unmarshal(resp.Body, &info)

	assert.Contains(t, info, "version")
}
