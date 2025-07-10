package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/reductstore/reductstore/pkg/models"
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
		log.DefaultLogger.Debug("Querying", "entry", qm.Entry, "from", from, "to", to)
		if from.After(to) {
			return &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					q.RefID: backend.ErrDataResponse(backend.StatusBadRequest, "from time is after to time"),
				},
			}, nil
		}
		options := reductgo.NewQueryOptionsBuilder().WithHead(true)
		if !from.IsZero() {
			options.WithStart(from.UnixMicro())
		}
		if !to.IsZero() {
			options.WithStop(to.UnixMicro())
		}
		res := d.query(ctx, req.PluginContext, qm.Bucket, qm.Entry, options.Build())
		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *ReductDatasource) query(ctx context.Context, pCtx backend.PluginContext, bucketName string, entry string, options reductgo.QueryOptions) backend.DataResponse {

	// Call your SDK
	bucket, err := d.reductClient.GetBucket(ctx, bucketName)
	if err != nil {
		log.DefaultLogger.Error("Failed to get bucket", "error", err)
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("get bucket: %v", err.Error()))
	}
	records, err := bucket.Query(ctx, entry, &options)
	if err != nil {
		log.DefaultLogger.Error("Failed to query", "error", err)
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("query: %v", err.Error()))
	}

	// Build a time series frame
	times := make([]time.Time, 0)
	contentTypes := make([]string, 0)
	labelFieldsInt64 := make(map[string][]*int64)
	labelFieldsFloat64 := make(map[string][]*float64)
	labelFieldsBool := make(map[string][]*bool)
	labelFieldsString := make(map[string][]*string)
	labelFieldsTime := make(map[string][]*time.Time)
	labelFieldsJson := make(map[string][]json.RawMessage)
	labelTypes := make(map[string]string) // Track the type of each label

	for record := range records.Records() {
		for key, value := range record.Labels() {
			strValue := fmt.Sprintf("%v", value)
			if _, exists := labelTypes[key]; !exists {
				// Try to determine type from first value
				if v, err := strconv.ParseInt(strValue, 10, 64); err == nil {
					labelTypes[key] = "int64"
					labelFieldsInt64[key] = make([]*int64, len(times))
					val := v
					labelFieldsInt64[key] = append(labelFieldsInt64[key], &val)
				} else if v, err := strconv.ParseFloat(strValue, 64); err == nil {
					labelTypes[key] = "float64"
					labelFieldsFloat64[key] = make([]*float64, len(times))
					val := v
					labelFieldsFloat64[key] = append(labelFieldsFloat64[key], &val)
				} else if strings.EqualFold(strValue, "true") || strings.EqualFold(strValue, "false") {
					labelTypes[key] = "bool"
					labelFieldsBool[key] = make([]*bool, len(times))
					val := strings.EqualFold(strValue, "true")
					labelFieldsBool[key] = append(labelFieldsBool[key], &val)
				} else if t, err := time.Parse(time.RFC3339, strValue); err == nil {
					labelTypes[key] = "time"
					labelFieldsTime[key] = make([]*time.Time, len(times))
					labelFieldsTime[key] = append(labelFieldsTime[key], &t)
				} else if json.Valid([]byte(strValue)) {
					labelTypes[key] = "json"
					labelFieldsJson[key] = make([]json.RawMessage, len(times))
					labelFieldsJson[key] = append(labelFieldsJson[key], json.RawMessage(strValue))
				} else {
					labelTypes[key] = "string"
					labelFieldsString[key] = make([]*string, len(times))
					labelFieldsString[key] = append(labelFieldsString[key], &strValue)
				}
			} else {
				// Convert value based on established type
				switch labelTypes[key] {
				case "int64":
					if v, err := strconv.ParseInt(strValue, 10, 64); err == nil {
						labelFieldsInt64[key] = append(labelFieldsInt64[key], &v)
					} else {
						v := int64(0)
						labelFieldsInt64[key] = append(labelFieldsInt64[key], &v)
					}
				case "float64":
					if v, err := strconv.ParseFloat(strValue, 64); err == nil {
						labelFieldsFloat64[key] = append(labelFieldsFloat64[key], &v)
					} else {
						v := float64(0)
						labelFieldsFloat64[key] = append(labelFieldsFloat64[key], &v)
					}
				case "bool":
					v := strings.EqualFold(strValue, "true")
					labelFieldsBool[key] = append(labelFieldsBool[key], &v)
				case "time":
					if t, err := time.Parse(time.RFC3339, strValue); err == nil {
						labelFieldsTime[key] = append(labelFieldsTime[key], &t)
					} else {
						t := time.Time{}
						labelFieldsTime[key] = append(labelFieldsTime[key], &t)
					}
				case "json":
					labelFieldsJson[key] = append(labelFieldsJson[key], json.RawMessage(strValue))
				default:
					labelFieldsString[key] = append(labelFieldsString[key], &strValue)
				}
			}
		}
		times = append(times, time.Unix(record.Time(), 0))
		contentTypes = append(contentTypes, record.ContentType())
	}

	// determine the max length
	maxLength := len(times)
	maxLength = int(math.Max(float64(maxLength), float64(len(contentTypes))))

	// pad only remaining fields with empty values
	if len(times) < maxLength {
		previousTime := times[len(times)-1]
		times = append(times, make([]time.Time, maxLength-len(times))...)
		for i := len(times) - 1; i >= len(times)-(maxLength-len(times)); i-- {
			times[i] = previousTime
			previousTime = previousTime.Add(time.Second)
		}
	}
	if len(contentTypes) < maxLength {
		previousContentType := contentTypes[len(contentTypes)-1]
		contentTypes = append(contentTypes, make([]string, maxLength-len(contentTypes))...)
		for i := len(contentTypes) - 1; i >= len(contentTypes)-(maxLength-len(contentTypes)); i-- {
			contentTypes[i] = previousContentType
		}
	}

	// Create the frame with time and content type
	frame := data.NewFrame("Entry Data",
		data.NewField("timestamp", nil, times),
		data.NewField("content_type", nil, contentTypes),
	)

	// Add label fields with proper padding
	for key, fieldType := range labelTypes {
		switch fieldType {
		case "int64":
			values := labelFieldsInt64[key]
			if len(values) < maxLength {
				padding := make([]*int64, maxLength-len(values))
				for i := range padding {
					v := int64(0)
					padding[i] = &v
				}
				values = append(values, padding...)
			}
			frame.Fields = append(frame.Fields, data.NewField(key, nil, values))
		case "float64":
			values := labelFieldsFloat64[key]
			if len(values) < maxLength {
				padding := make([]*float64, maxLength-len(values))
				for i := range padding {
					v := float64(0)
					padding[i] = &v
				}
				values = append(values, padding...)
			}
			frame.Fields = append(frame.Fields, data.NewField(key, nil, values))
		case "bool":
			values := labelFieldsBool[key]
			if len(values) < maxLength {
				padding := make([]*bool, maxLength-len(values))
				for i := range padding {
					v := false
					padding[i] = &v
				}
				values = append(values, padding...)
			}
			frame.Fields = append(frame.Fields, data.NewField(key, nil, values))
		case "time":
			values := labelFieldsTime[key]
			if len(values) < maxLength {
				padding := make([]*time.Time, maxLength-len(values))
				for i := range padding {
					t := time.Time{}
					padding[i] = &t
				}
				values = append(values, padding...)
			}
			frame.Fields = append(frame.Fields, data.NewField(key, nil, values))
		default:
			values := labelFieldsString[key]
			if len(values) < maxLength {
				padding := make([]*string, maxLength-len(values))
				for i := range padding {
					s := ""
					padding[i] = &s
				}
				values = append(values, padding...)
			}
			frame.Fields = append(frame.Fields, data.NewField(key, nil, values))
		}
	}

	return backend.DataResponse{
		Frames: []*data.Frame{frame},
	}
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
			Body:   []byte(fmt.Sprintf("unknown resource path: %s", req.Path)),
		})
	}
}

type reductQuery struct {
	Bucket  string                `json:"bucket"`
	Entry   string                `json:"entry"`
	Options reductgo.QueryOptions `json:"options"`
}

func (d *ReductDatasource) handleListBuckets(ctx context.Context, sender backend.CallResourceResponseSender) error {
	buckets, err := d.reductClient.GetBuckets(ctx)
	if err != nil {
		log.DefaultLogger.Error("Failed to get buckets", "error", err)
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(fmt.Sprintf("error: %v", err)),
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
			Body:   []byte(fmt.Sprintf("error getting bucket: %v", err)),
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
