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
	frames := getFramesV1(10, records.Records())
	// frames := getFramesV2(records.Records())
	// frames := getFramesV3(10, records.Records())
	// frames := getFramesV4(10, records.Records())
	// frames := getFramesV5(10, records.Records())
	return backend.DataResponse{
		Frames: frames,
	}
}
func getFramesV4(limit int, records <-chan *reductgo.ReadableRecord) []*data.Frame {
	frames := []*data.Frame{}
	start := time.Now()
	// Build a time series frame
	mainFrame := data.NewFrame("Entry Data")
	mainFrame.Meta = &data.FrameMeta{
		Type: data.FrameTypeTimeSeriesMulti,
	}
	contentTypeTrack := make(map[string]bool)
	for record := range records {
		for key, value := range record.Labels() {
			strValue := fmt.Sprintf("%v", value)
			// Try to determine type from first value
			labelFrame := data.NewFrame(fmt.Sprintf("Label: %s", key))
			labelFrame.Meta = &data.FrameMeta{
				Type: data.FrameTypeTimeSeriesMulti,
			}
			if v, err := strconv.ParseInt(strValue, 10, 64); err == nil {
				// add time field first
				labelFrame.Fields = append(labelFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []int64{v}))
			} else if v, err := strconv.ParseFloat(strValue, 64); err == nil {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []float64{v}))
			} else if strings.EqualFold(strValue, "true") || strings.EqualFold(strValue, "false") {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []bool{strings.EqualFold(strValue, "true")}))
			} else if t, err := time.Parse(time.RFC3339, strValue); err == nil {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []time.Time{t}))
			} else if json.Valid([]byte(strValue)) {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []json.RawMessage{json.RawMessage(strValue)}))
			} else {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []string{strValue}))
			}

			frames = append(frames, labelFrame)
		}
		mainFrame.Fields = append(mainFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
		if _, exists := contentTypeTrack[record.ContentType()]; !exists {
			mainFrame.Fields = append(mainFrame.Fields, data.NewField("content_type", nil, []string{record.ContentType()}))
			contentTypeTrack[record.ContentType()] = true
		}
		frames = append(frames, mainFrame)
		tsSchema := mainFrame.TimeSeriesSchema()
		if tsSchema.Type == data.TimeSeriesTypeLong {
			wideFrame, err := data.LongToWide(mainFrame, nil)
			if err == nil {
				frames = append(frames, wideFrame)
			} else {
				log.DefaultLogger.Debug("Failed to convert to wide frame", "error", err)
			}
		}

	}
	log.DefaultLogger.Debug("Completed fetching records", "took", time.Since(start))

	return frames
}

func getFramesV5(limit int, records <-chan *reductgo.ReadableRecord) []*data.Frame {
	frames := make(map[string]*data.Frame)
	start := time.Now()

	// Maps to store values for each label
	labelTimes := make(map[string][]time.Time)
	labelValues := make(map[string]interface{})
	labelTypes := make(map[string]string)

	// First pass: determine types and collect values
	for record := range records {
		for key, value := range record.Labels() {
			strValue := fmt.Sprintf("%v", value)
			timestamp := time.UnixMicro(record.Time())

			// Initialize arrays if this is a new label
			if _, exists := frames[key]; !exists {
				frames[key] = data.NewFrame(fmt.Sprintf("Label: %s", key))
				frames[key].Meta = &data.FrameMeta{
					Type: data.FrameTypeTimeSeriesWide,
				}
				labelTimes[key] = make([]time.Time, 0)
			}

			// Determine type if not already determined
			if _, exists := labelTypes[key]; !exists {
				// Try to determine type from first value
				if _, err := strconv.ParseInt(strValue, 10, 64); err == nil {
					labelTypes[key] = "int64"
					labelValues[key] = make([]int64, 0)
				} else if _, err := strconv.ParseFloat(strValue, 64); err == nil {
					labelTypes[key] = "float64"
					labelValues[key] = make([]float64, 0)
				} else if strings.EqualFold(strValue, "true") || strings.EqualFold(strValue, "false") {
					labelTypes[key] = "bool"
					labelValues[key] = make([]bool, 0)
				} else if _, err := time.Parse(time.RFC3339, strValue); err == nil {
					labelTypes[key] = "time"
					labelValues[key] = make([]time.Time, 0)
				} else if json.Valid([]byte(strValue)) {
					labelTypes[key] = "json"
					labelValues[key] = make([]json.RawMessage, 0)
				} else {
					labelTypes[key] = "string"
					labelValues[key] = make([]string, 0)
				}
			}

			// Append timestamp
			labelTimes[key] = append(labelTimes[key], timestamp)

			// Append value based on type
			switch labelTypes[key] {
			case "int64":
				if v, err := strconv.ParseInt(strValue, 10, 64); err == nil {
					labelValues[key] = append(labelValues[key].([]int64), v)
				} else {
					labelValues[key] = append(labelValues[key].([]int64), 0)
				}
			case "float64":
				if v, err := strconv.ParseFloat(strValue, 64); err == nil {
					labelValues[key] = append(labelValues[key].([]float64), v)
				} else {
					labelValues[key] = append(labelValues[key].([]float64), 0.0)
				}
			case "bool":
				labelValues[key] = append(labelValues[key].([]bool), strings.EqualFold(strValue, "true"))
			case "time":
				if t, err := time.Parse(time.RFC3339, strValue); err == nil {
					labelValues[key] = append(labelValues[key].([]time.Time), t)
				} else {
					labelValues[key] = append(labelValues[key].([]time.Time), time.Time{})
				}
			case "json":
				if json.Valid([]byte(strValue)) {
					labelValues[key] = append(labelValues[key].([]json.RawMessage), json.RawMessage(strValue))
				} else {
					labelValues[key] = append(labelValues[key].([]json.RawMessage), json.RawMessage("{}"))
				}
			default: // string
				labelValues[key] = append(labelValues[key].([]string), strValue)
			}
		}
	}

	// Create frames from collected data
	for key := range frames {
		// Add timestamp field
		frames[key].Fields = append(frames[key].Fields,
			data.NewField("timestamp", nil, labelTimes[key]))

		// Add value field based on type
		switch labelTypes[key] {
		case "int64":
			frames[key].Fields = append(frames[key].Fields,
				data.NewField(key, nil, labelValues[key].([]int64)))
		case "float64":
			frames[key].Fields = append(frames[key].Fields,
				data.NewField(key, nil, labelValues[key].([]float64)))
		case "bool":
			frames[key].Fields = append(frames[key].Fields,
				data.NewField(key, nil, labelValues[key].([]bool)))
		case "time":
			frames[key].Fields = append(frames[key].Fields,
				data.NewField(key, nil, labelValues[key].([]time.Time)))
		case "json":
			frames[key].Fields = append(frames[key].Fields,
				data.NewField(key, nil, labelValues[key].([]json.RawMessage)))
		default: // string
			frames[key].Fields = append(frames[key].Fields,
				data.NewField(key, nil, labelValues[key].([]string)))
		}
	}

	log.DefaultLogger.Debug("Completed fetching records", "took", time.Since(start))

	// Convert map to slice
	result := make([]*data.Frame, 0, len(frames))
	for _, frame := range frames {
		tsSchema := frame.TimeSeriesSchema()
		if tsSchema.Type == data.TimeSeriesTypeLong {
			wideFrame, err := data.LongToWide(frame, nil)
			if err == nil {
				result = append(result, wideFrame)
			} else {
				log.DefaultLogger.Debug("Failed to convert to wide frame", "error", err)
			}
		} else {
			result = append(result, frame)
		}
	}

	return result
}
func getFramesV3(limit int, records <-chan *reductgo.ReadableRecord) []*data.Frame {
	start := time.Now()
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

	for record := range records {
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
	tsSchema := frame.TimeSeriesSchema()
	if tsSchema.Type == data.TimeSeriesTypeLong {
		wideFrame, err := data.LongToWide(frame, nil)
		if err == nil {
			log.DefaultLogger.Debug("Completed converting to wide frame", "took", time.Since(start))
			return []*data.Frame{wideFrame}
		} else {
			log.DefaultLogger.Debug("Failed to convert to wide frame", "error", err)
		}
	}
	log.DefaultLogger.Debug("Completed fetching records", "took", time.Since(start))
	return []*data.Frame{frame} // return the frame as is
}
func getFramesV2(records <-chan *reductgo.ReadableRecord) []*data.Frame {
	frames := []*data.Frame{}
	start := time.Now()
	// Build a time series frame
	mainFrame := data.NewFrame("Entry Data")
	mainFrame.Meta = &data.FrameMeta{
		Type: data.FrameTypeTimeSeriesWide,
	}
	times := make(map[time.Time]bool)
	contentTypes := make(map[string]bool)
	countLabels := make(map[string]int64)
	for record := range records {
		for key, value := range record.Labels() {
			strValue := fmt.Sprintf("%v", value)
			countLabels[key]++
			// Try to determine type from first value
			labelFrame := data.NewFrame(fmt.Sprintf("Label: %s", key))
			labelFrame.Meta = &data.FrameMeta{
				Type: data.FrameTypeTimeSeriesWide,
			}
			labelFrame.Fields = append(labelFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
			if v, err := strconv.ParseInt(strValue, 10, 64); err == nil {
				// add time field first
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []int64{v}))
			} else if v, err := strconv.ParseFloat(strValue, 64); err == nil {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []float64{v}))
			} else if strings.EqualFold(strValue, "true") || strings.EqualFold(strValue, "false") {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []bool{strings.EqualFold(strValue, "true")}))
			} else if t, err := time.Parse(time.RFC3339, strValue); err == nil {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []time.Time{t}))
			} else if json.Valid([]byte(strValue)) {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []json.RawMessage{json.RawMessage(strValue)}))
			} else {
				labelFrame.Fields = append(labelFrame.Fields, data.NewField(key, nil, []string{strValue}))
			}
			tsLabelFrameSchema := labelFrame.TimeSeriesSchema()
			if tsLabelFrameSchema.Type == data.TimeSeriesTypeLong {
				wideLabelFrame, err := data.LongToWide(labelFrame, nil)
				if err == nil {
					frames = append(frames, wideLabelFrame)
				} else {
					log.DefaultLogger.Debug("Failed to convert to wide frame", "error", err)
				}
			} else {
				log.DefaultLogger.Debug("Label frame is not a long frame", "type", tsLabelFrameSchema.Type)
			}
		}
		exists := times[time.UnixMicro(record.Time())]
		if !exists {
			times[time.UnixMicro(record.Time())] = true
			mainFrame.Fields = append(mainFrame.Fields, data.NewField("timestamp", nil, []time.Time{time.UnixMicro(record.Time())}))
		}
		exists = contentTypes[record.ContentType()]
		if !exists {
			contentTypes[record.ContentType()] = true
			mainFrame.Fields = append(mainFrame.Fields, data.NewField("content_type", nil, []string{record.ContentType()}))
		}
	}
	tsSchema := mainFrame.TimeSeriesSchema()
	if tsSchema.Type == data.TimeSeriesTypeLong {
		wideFrame, err := data.LongToWide(mainFrame, nil)
		if err == nil {
			frames = append(frames, wideFrame)
		} else {
			log.DefaultLogger.Debug("Failed to convert to wide frame", "error", err)
		}
	} else {
		log.DefaultLogger.Debug("Main frame is not a long frame", "type", tsSchema.Type)
	}

	log.DefaultLogger.Debug("Completed fetching records", "took", time.Since(start))
	return frames
}

func getFramesV1(limit int, records <-chan *reductgo.ReadableRecord) []*data.Frame {
	frames := []*data.Frame{}
	start := time.Now()
	// Build a time series frame
	mainFrame := data.NewFrame("Entry Data")
	mainFrame.Meta = &data.FrameMeta{
		Type:                   data.FrameTypeTimeSeriesWide,
		PreferredVisualization: data.VisTypeTable,
	}
	times := make([]time.Time, 0)
	contentTypes := make([]string, 0)
	labelCounts := make(map[string]int8)
	contentTypesMap := make(map[string]int32)
	labels := make([]string, 0)
	values := make([]string, 0)
	for record := range records {
		for key, value := range record.Labels() {
			strValue := fmt.Sprintf("%v", value)
			// Try to determine type from first value
			labels = append(labels, key)
			labelCounts[key]++
			values = append(values, strValue)
			contentTypesMap[record.ContentType()]++
			contentTypes = append(contentTypes, record.ContentType())
			times = append(times, time.UnixMicro(record.Time()))
		}
	}

	mainFrame.Fields = append(mainFrame.Fields, data.NewField("timestamp", nil, times[:limit]).SetConfig(&data.FieldConfig{
		Interval: float64(time.Microsecond),
	}))
	mainFrame.Fields = append(mainFrame.Fields, data.NewField("content_type", nil, contentTypes[:limit]))
	countFields := make([]int32, 0)
	for _, contentType := range contentTypes {
		countFields = append(countFields, contentTypesMap[contentType])
	}
	mainFrame.Fields = append(mainFrame.Fields, data.NewField("count", nil, countFields[:limit]))

	mainFrame.Fields = append(mainFrame.Fields, data.NewField("label", nil, labels[:limit]))

	mainFrame.Fields = append(mainFrame.Fields, data.NewField("value", nil, values[:limit]))
	tsSchema := mainFrame.TimeSeriesSchema()
	if tsSchema.Type == data.TimeSeriesTypeLong {
		wideFrame, err := data.LongToWide(mainFrame, &data.FillMissing{
			Mode: data.FillModeNull,
		})
		if err == nil {
			frames = append(frames, wideFrame)
		} else {
			log.DefaultLogger.Debug("Failed to convert to wide frame", "error", err)
		}
	}
	log.DefaultLogger.Debug("Completed fetching records", "took", time.Since(start))
	return frames
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
