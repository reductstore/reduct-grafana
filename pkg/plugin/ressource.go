package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/reductstore/reduct-go/model"
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

	case "validateCondition":
		log.DefaultLogger.Debug("Received validateCondition", "body", req.Body)
		return d.handleValidateCondition(ctx, req, sender)

	case "serverInfo":
		log.DefaultLogger.Debug("Received serverInfo")
		return d.handleServerInfo(ctx, sender)

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
		errorResp := map[string]string{"error": fmt.Sprintf("error: %v", err)}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	resp, err := json.Marshal(buckets)
	if err != nil {
		errorResp := map[string]string{"error": "failed to marshal bucket list"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
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
		errorResp := map[string]string{"error": "invalid request body"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadRequest,
			Body:   errorJson,
		})
	}

	err = json.Unmarshal(body, &payload)
	if err != nil || payload.Bucket == "" {
		log.DefaultLogger.Warn("Missing or invalid bucket in request")
		errorResp := map[string]string{"error": "missing or invalid 'bucket' in request"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadRequest,
			Body:   errorJson,
		})
	}

	bucket, err := d.reductClient.GetBucket(ctx, payload.Bucket)
	if err != nil {
		log.DefaultLogger.Error("Failed to get bucket", "bucket", payload.Bucket, "error", err)
		errorResp := map[string]string{"error": fmt.Sprintf("error getting bucket: %v", err)}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	entries, err := bucket.GetEntries(ctx)
	if err != nil {
		log.DefaultLogger.Error("Failed to list entries", "error", err)
		errorResp := map[string]string{"error": "error getting entries"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	resp, err := json.Marshal(entries)
	if err != nil {
		errorResp := map[string]string{"error": "failed to marshal entries"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   resp,
	})
}

func (d *ReductDatasource) handleValidateCondition(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	var payload struct {
		Bucket    string `json:"bucket"`
		Entry     string `json:"entry"`
		Condition any    `json:"condition"`
	}

	body, err := io.ReadAll(bytes.NewReader(req.Body))
	if err != nil {
		log.DefaultLogger.Error("Failed to read request body", "error", err)
		errorResp := map[string]string{"error": "invalid request body"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadRequest,
			Body:   errorJson,
		})
	}

	err = json.Unmarshal(body, &payload)
	if err != nil || payload.Bucket == "" || payload.Entry == "" {
		log.DefaultLogger.Warn("Missing or invalid bucket/entry in request")
		errorResp := map[string]string{"error": "missing or invalid 'bucket' or 'entry' in request"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusBadRequest,
			Body:   errorJson,
		})
	}

	// Create modified condition with $limit: 1 for validation
	conditionMap, err := parseAndNormalizeCondition(payload.Condition)
	if err != nil {
		response := map[string]any{
			"valid": false,
			"error": err.Error(),
		}
		resp, _ := json.Marshal(response)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   resp,
		})
	}

	// Add or override $limit to 1 for validation
	conditionMap["$limit"] = 1
	modifiedCondition := conditionMap

	// Test the condition by running a query with it
	bucket, err := d.reductClient.GetBucket(ctx, payload.Bucket)
	if err != nil {
		log.DefaultLogger.Error("Failed to get bucket", "bucket", payload.Bucket, "error", err)
		var apiErr model.APIError
		errorMsg := fmt.Sprintf("Failed to access bucket '%s'", payload.Bucket)
		if errors.As(err, &apiErr) {
			errorMsg = apiErr.Message
		}
		response := map[string]any{
			"valid": false,
			"error": errorMsg,
		}
		resp, _ := json.Marshal(response)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   resp,
		})
	}

	// Create query options with the modified condition
	options := reductgo.NewQueryOptionsBuilder().WithWhen(modifiedCondition).Build()

	// Try to execute the query to validate the condition
	records, err := bucket.Query(ctx, payload.Entry, &options)
	if err != nil {
		log.DefaultLogger.Debug("Query validation failed", "error", err)
		var apiErr model.APIError
		errorMsg := "Query validation failed"
		if errors.As(err, &apiErr) {
			errorMsg = apiErr.Message
		}
		response := map[string]any{
			"valid": false,
			"error": errorMsg,
		}
		resp, _ := json.Marshal(response)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Body:   resp,
		})
	}

	// Close the records channel to clean up
	go func() {
		for range records.Records() {
			// Drain the channel
		}
	}()

	// If we get here, the condition is valid
	response := map[string]any{
		"valid": true,
	}
	resp, err := json.Marshal(response)
	if err != nil {
		errorResp := map[string]string{"error": "failed to marshal validation response"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   resp,
	})
}

func parseAndNormalizeCondition(condition any) (map[string]any, error) {
	// Default interval used for validating the condition
	const defaultInterval = "1s"

	if condition == nil {
		return map[string]any{}, nil
	}

	switch v := condition.(type) {
	case map[string]any:
		return replaceIntervalMacros(v, defaultInterval).(map[string]any), nil
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return map[string]any{}, nil
		}

		var parsed any
		err := json.Unmarshal([]byte(trimmed), &parsed)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON syntax: %v", err)
		}

		parsedMap, ok := parsed.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid JSON syntax: expected object for condition")
		}

		return replaceIntervalMacros(parsedMap, defaultInterval).(map[string]any), nil
	default:
		return nil, fmt.Errorf("invalid condition: expected object or JSON string")
	}
}

func replaceIntervalMacros(value any, interval string) any {
	switch v := value.(type) {
	case string:
		if v == "$__interval" {
			return interval
		}
		return v
	case []any:
		for i := range v {
			v[i] = replaceIntervalMacros(v[i], interval)
		}
		return v
	case map[string]any:
		for key, val := range v {
			v[key] = replaceIntervalMacros(val, interval)
		}
		return v
	default:
		return value
	}
}

func (d *ReductDatasource) handleServerInfo(ctx context.Context, sender backend.CallResourceResponseSender) error {
	// Check if server is live first
	_, err := d.reductClient.IsLive(ctx)
	if err != nil {
		log.DefaultLogger.Error("Failed to check server status", "error", err)
		errorResp := map[string]string{"error": "server is not accessible"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	// Get server info
	serverInfo, err := d.reductClient.GetInfo(ctx)
	if err != nil {
		log.DefaultLogger.Error("Failed to get server info", "error", err)
		errorResp := map[string]string{"error": "failed to get server info"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	resp, err := json.Marshal(serverInfo)
	if err != nil {
		errorResp := map[string]string{"error": "failed to marshal server info"}
		errorJson, _ := json.Marshal(errorResp)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   errorJson,
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   resp,
	})
}
