package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	reductgo "github.com/reductstore/reduct-go"
	model "github.com/reductstore/reduct-go/model"
)

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *ReductDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		var qm reductQuery

		err := json.Unmarshal(q.JSON, &qm)
		if err != nil {
			log.DefaultLogger.Error("Failed to unmarshal query", "error", err)
			return &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					q.RefID: backend.ErrDataResponse(backend.StatusBadRequest, "invalid query format"),
				},
			}, nil
		}

		log.DefaultLogger.Debug(
			"QueryData received",
			"ref_id", q.RefID,
			"bucket", qm.Bucket,
			"entry", qm.Entry,
			"mode", qm.Options.Mode,
			"from", q.TimeRange.From.UTC(),
			"to", q.TimeRange.To.UTC(),
		)

		// Validate required fields
		if qm.Bucket == "" || qm.Entry == "" {
			return &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					q.RefID: backend.ErrDataResponse(backend.StatusBadRequest, "missing bucket or entry"),
				},
			}, nil
		}
		from := q.TimeRange.From.UTC()
		to := q.TimeRange.To.UTC()

		when := qm.Options.When
		mode := qm.Options.Mode

		if from.After(to) {
			return &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					q.RefID: backend.ErrDataResponse(backend.StatusBadRequest, "from time is after to time"),
				},
			}, nil
		}

		options := reductgo.NewQueryOptionsBuilder().WithWhen(when)
		if mode == ModeLabelOnly {
			options.WithHead(true)
		} else {
			options.WithHead(false)
		}

		if !from.IsZero() {
			options.WithStart(from.UnixMicro())
		}
		if !to.IsZero() {
			options.WithStop(to.UnixMicro())
		}
		res := d.query(ctx, req.PluginContext, qm.Bucket, qm.Entry, options.Build(), mode)
		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *ReductDatasource) query(
	ctx context.Context,
	pCtx backend.PluginContext,
	bucketName string,
	entry string,
	options reductgo.QueryOptions,
	mode ReductMode,
) backend.DataResponse {
	bucket, err := d.reductClient.GetBucket(ctx, bucketName)
	if err != nil {
		log.DefaultLogger.Error("Failed to get bucket", "error", err)
		var apiErr model.APIError
		errors.As(err, &apiErr)
		return backend.ErrDataResponse(backend.Status(apiErr.Status), apiErr.Message)
	}
	records, err := bucket.Query(ctx, entry, &options)
	if err != nil {
		log.DefaultLogger.Error("Failed to query", "error", err)
		var apiErr model.APIError
		errors.As(err, &apiErr)
		return backend.ErrDataResponse(backend.Status(apiErr.Status), apiErr.Message)
	}

	frames := getFrames(records.Records(), mode)
	return backend.DataResponse{
		Frames: frames,
	}
}

func getFrames(records <-chan *reductgo.ReadableRecord, mode ReductMode) []*data.Frame {
	frames := make(map[string]*data.Frame)
	labelKinds := make(map[string]reflect.Kind)

	for record := range records {
		if mode == "" || mode == ModeLabelOnly || mode == ModeLabelAndContent {
			processLabels(frames, labelKinds, record)
		}
		if mode == ModeContentOnly || mode == ModeLabelAndContent {
			processContent(frames, record)
		}
	}

	result := make([]*data.Frame, 0, len(frames))
	for _, frame := range frames {
		// Append timestamp field if not already present
		result = append(result, frame)
	}
	return result
}

// processContent processes the content of a record and appends it to the frames.
func processLabels(frames map[string]*data.Frame, kindMap map[string]reflect.Kind, record *reductgo.ReadableRecord) {
	for key, labelValue := range record.Labels() {

		strValue := fmt.Sprintf("%v", labelValue)
		initialType, ok := kindMap[key]
		value := parseValue(strValue)

		if !ok {
			kind := reflect.TypeOf(value).Kind()
			kindMap[key] = kind
			initialType = kind
		}

		currentType := reflect.TypeOf(value).Kind()
		if currentType != initialType {
			// If the type has changed, we need to coerce the value to the initial type
			log.DefaultLogger.Debug("Type change detected", "key", key, "from", initialType, "to", currentType)
			val, err := coerceToKind(strValue, initialType)
			if err != nil {
				log.DefaultLogger.Error("Failed to coerce value", "key", key, "value", strValue, "error", err)
				continue
			}
			value = val
		}

		switch v := value.(type) {
		case int64:
			appendValue(frames, key, record, v)
		case float64:
			appendValue(frames, key, record, v)
		case bool:
			appendValue(frames, key, record, v)
		case string:
			appendValue(frames, key, record, v)
		default:
			appendValue(frames, key, record, strValue)
		}
	}
}

// processContent reads record body, parses JSON, flattens it, and appends values to frames.
func processContent(
	frames map[string]*data.Frame,
	record *reductgo.ReadableRecord,
) {
	s, err := record.ReadAsString()
	if err != nil || len(strings.TrimSpace(s)) == 0 {
		return
	}

	b := []byte(s)
	if !looksLikeJSON(b) {
		return
	}

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return
	}

	flat := map[string]any{}
	flattenJSON("$", v, flat)

	for k, val := range flat {
		switch v := val.(type) {
		case int64:
			appendValue(frames, k, record, v)
		case float64:
			appendValue(frames, k, record, v)
		case bool:
			appendValue(frames, k, record, v)
		case string:
			appendValue(frames, k, record, v)
		default:
			str := fmt.Sprintf("%v", val)
			appendValue(frames, k, record, str)
		}
	}
}

func looksLikeJSON(b []byte) bool {
	for _, c := range b {
		if c == ' ' || c == '\n' || c == '\t' || c == '\r' {
			continue
		}
		return c == '{' || c == '['
	}
	return false
}

func flattenJSON(prefix string, v any, out map[string]any) {
	switch t := v.(type) {
	case map[string]any:
		for k, vv := range t {
			flattenJSON(prefix+"."+k, vv, out)
		}
	case []any:
		for i, vv := range t {
			flattenJSON(fmt.Sprintf("%s[%d]", prefix, i), vv, out)
		}
	default:
		out[prefix] = v
	}
}

// appendValue appends a value to the frame for the given key.
func appendValue[V float64 | int64 | bool | string](frames map[string]*data.Frame, key string, record *reductgo.ReadableRecord, val V) {
	// Check if frame for this label already exists
	if frame, exists := frames[key]; exists {
		// Append new value to existing frame
		frame.Fields[0].Append(time.UnixMicro(record.Time()))
		frame.Fields[1].Append(val)
	} else {
		// Create a new frame for this label
		frame = data.NewFrame(key,
			data.NewField("time", nil, []time.Time{time.UnixMicro(record.Time())}),
			data.NewField("value", nil, []V{val}),
		)

		frame.Meta = &data.FrameMeta{
			Type: data.FrameTypeTimeSeriesWide,
		}
		frames[key] = frame
	}
}

// coerceToKind attempts to convert a string value to the specified reflect.Kind type.
func coerceToKind(str string, kind reflect.Kind) (any, error) {
	switch kind {
	case reflect.Int, reflect.Int64:
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return int64(f), nil
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f, nil
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(str); err == nil {
			return b, nil
		}
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f != 0, nil
		}
		return false, fmt.Errorf("invalid boolean value")
	case reflect.String:
		return str, nil
	default:
		return str, fmt.Errorf("unsupported kind: %s", kind)
	}

	return str, fmt.Errorf("coerceToKind: failed to coerce value '%s' to kind '%s'", str, kind)
}

// parseValue parses a string value into the appropriate type based on the kind.
func parseValue(str string) any {
	if v, err := strconv.ParseInt(str, 10, 64); err == nil {
		return v
	}

	if v, err := strconv.ParseFloat(str, 64); err == nil {
		return v
	}

	if v, err := strconv.ParseBool(str); err == nil {
		return v
	}

	// Default to string if no other type matches
	return str
}
