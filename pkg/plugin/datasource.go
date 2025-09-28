package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/reductstore/reductstore/pkg/models"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	reductgo "github.com/reductstore/reduct-go"
	model "github.com/reductstore/reduct-go/model"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*ReductDatasource)(nil)
	_ backend.CheckHealthHandler    = (*ReductDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*ReductDatasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// Get the URL and API token from JSON config
	pluginSettings, err := models.LoadPluginSettings(settings)
	if err != nil {
		log.DefaultLogger.Error("load plugin settings", "error", err)
		return nil, fmt.Errorf("load plugin settings: %w", err)
	}
	// check both server url and server token are in the plugin settings
	if pluginSettings.ServerURL == "" {
		log.DefaultLogger.Error("server URL is missing")
		return nil, fmt.Errorf("server URL is missing")
	}
	client := reductgo.NewClient(pluginSettings.ServerURL, reductgo.ClientOptions{
		APIToken:  pluginSettings.Secrets.ServerToken,
		VerifySSL: pluginSettings.VerifySSL,
	})
	_, err = client.IsLive(ctx)
	if err != nil {
		log.DefaultLogger.Error("check health failed", "error", err)
		return nil, fmt.Errorf("check health failed: %w", err)
	}

	return &ReductDatasource{reductClient: client}, nil
}

// ReductDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type ReductDatasource struct {
	reductClient reductgo.Client
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *ReductDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *ReductDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()
	log.DefaultLogger.Debug("Received QueryData", "queries", req.Queries)

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

		log.DefaultLogger.Debug(
			"Querying",
			"entry", qm.Entry,
			"from", from, "to", to,
			"when", when,
			"mode", mode,
		)
		if from.After(to) {
			return &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					q.RefID: backend.ErrDataResponse(backend.StatusBadRequest, "from time is after to time"),
				},
			}, nil
		}
	
		options := reductgo.NewQueryOptionsBuilder().WithWhen(when)
		if mode == ModeLabels {
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

func (d *ReductDatasource) query(ctx context.Context, pCtx backend.PluginContext, bucketName string, entry string, options reductgo.QueryOptions, mode ReductMode) backend.DataResponse {
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
	// map of frames for each label
	frames := make(map[string]*data.Frame)
	kinds  := make(map[string]reflect.Kind)

	for record := range records {
		if mode == "" || mode == ModeLabels || mode == ModeBoth {
			processLabels(frames, kinds, record)
		}
		if mode == ModeContent || mode == ModeBoth {
			processContent(frames, kinds, record)
		}
	}

	// return frames as an array
	result := make([]*data.Frame, 0, len(frames))
	for _, frame := range frames {
		// Append timestamp field if not already present
		result = append(result, frame)
	}
	return result
}

// processContent processes the content of a record and appends it to the frames.
func processLabels(frames map[string]*data.Frame, labelInitialType map[string]reflect.Kind, record *reductgo.ReadableRecord) {
	for key, labelValue := range record.Labels() {

		strValue := fmt.Sprintf("%v", labelValue)
		initialType, ok := labelInitialType[key]
		value := parseValue(strValue)

		if !ok {
			kind := reflect.TypeOf(value).Kind()
			labelInitialType[key] = kind
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

// processContent reads record body, tries to parse JSON, flattens it, and appends values as frames.
func processContent(
  frames map[string]*data.Frame,
  initial map[string]reflect.Kind,
  record *reductgo.ReadableRecord,
) {
  s, err := record.ReadAsString()
  if err != nil || len(strings.TrimSpace(s)) == 0 {
    return
  }

  // quick JSON sniff
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

  for k, raw := range flat {
    str := fmt.Sprintf("%v", raw)
    kind, ok := initial[k]
    val := parseValue(str)

    if !ok {
      kind = reflect.TypeOf(val).Kind()
      initial[k] = kind
    } else if reflect.TypeOf(val).Kind() != kind {
      if coerced, err := coerceToKind(str, kind); err == nil {
        val = coerced
      } else {
        val = str
      }
    }

    switch vv := val.(type) {
    case int64:
      appendValue(frames, k, record, vv)
    case float64:
      appendValue(frames, k, record, vv)
    case bool:
      appendValue(frames, k, record, vv)
    case string:
      appendValue(frames, k, record, vv)
    default:
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

	return str // Default to string if no other type matches
}

// CallResource handles HTTP resource requests from the frontend (e.g. dropdown fetching).
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

type ReductMode string

const (
  ModeLabels  ReductMode = "labels"
  ModeContent ReductMode = "content"
  ModeBoth    ReductMode = "both"
)

type reductOptions struct {
	Start        int64       `json:"start,omitempty"`
	Stop         int64       `json:"stop,omitempty"`
	When         any         `json:"when,omitempty"`
	Strict       bool        `json:"strict,omitempty"`
	Continuous   bool        `json:"continuous,omitempty"`
	Ext          any         `json:"ext,omitempty"`
	Mode         ReductMode  `json:"mode,omitempty"`
}

type reductQuery struct {
	Bucket string `json:"bucket"`
	Entry  string `json:"entry"`
	Options reductOptions  `json:"options"`
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

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *ReductDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	// url is not in secured json data, its in json data
	pluginSettings, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	// check for server url
	if pluginSettings.ServerURL == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Server URL is missing"
		return res, nil
	}
	// check for server token

	client := reductgo.NewClient(pluginSettings.ServerURL, reductgo.ClientOptions{
		APIToken:  pluginSettings.Secrets.ServerToken,
		VerifySSL: pluginSettings.VerifySSL,
	})
	_, err = client.IsLive(ctx)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to connect to server"
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
