package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func (d *ReductDatasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	log.DefaultLogger.Debug("Received CallResource", "Path", req.Path, "Method", req.Method)

	switch req.Path {
	case "listBuckets":
		log.DefaultLogger.Debug("Received listBuckets")
		return d.handleListBuckets(ctx, sender)

	case "listEntries":
		log.DefaultLogger.Debug("Received listEntries", "bucket", req.Body)
		return d.handleListEntries(ctx, req, sender)

	default:
		log.DefaultLogger.Warn("Unknown resource path", "path", req.Path)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusNotFound,
			Body:   fmt.Appendf(nil, "unknown resource path: %s", req.Path),
		})
	}
}

func (d *ReductDatasource) handleListBuckets(ctx context.Context, sender backend.CallResourceResponseSender) error {
	buckets, err := d.reductClient.GetBuckets(ctx)
	if err != nil {
		log.DefaultLogger.Error("Failed to get buckets", "error", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Appendf(nil, "error: %v", err),
		})
	}

	resp, err := json.Marshal(buckets)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("failed to marshal bucket list"),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   resp,
	})
}

func (d *ReductDatasource) handleListEntries(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	var payload struct {
		Bucket string `json:"bucket"`
	}

	body, err := io.ReadAll(bytes.NewReader(req.Body))
	if err != nil {
		log.DefaultLogger.Error("Failed to read request body", "error", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadRequest,
			Body:   []byte("invalid request body"),
		})
	}

	err = json.Unmarshal(body, &payload)
	if err != nil || payload.Bucket == "" {
		log.DefaultLogger.Warn("Missing or invalid bucket in request")
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadRequest,
			Body:   []byte("missing or invalid 'bucket' in request"),
		})
	}

	bucket, err := d.reductClient.GetBucket(ctx, payload.Bucket)
	if err != nil {
		log.DefaultLogger.Error("Failed to get bucket", "bucket", payload.Bucket, "error", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   fmt.Appendf(nil, "error getting bucket: %v", err),
		})
	}

	entries, err := bucket.GetEntries(ctx)
	if err != nil {
		log.DefaultLogger.Error("Failed to list entries", "error", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("error getting entries"),
		})
	}

	resp, err := json.Marshal(entries)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte("failed to marshal entries"),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   resp,
	})
}
