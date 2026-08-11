package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
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
			"entries", qm.Entries,
			"mode", qm.Options.Mode,
			"from", q.TimeRange.From.UTC(),
			"to", q.TimeRange.To.UTC(),
		)

		entries := qm.Entries
		if len(entries) == 0 && qm.Entry != "" {
			entries = []string{qm.Entry}
		}

		if qm.Bucket == "" || len(entries) == 0 {
			return &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					q.RefID: backend.ErrDataResponse(backend.StatusBadRequest, "missing bucket or entries"),
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
		res := d.query(ctx, req.PluginContext, qm.Bucket, entries, options.Build(), mode, qm.Options.CombinedFrame)
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
	entries []string,
	options reductgo.QueryOptions,
	mode ReductMode,
	combinedFrame bool,
) backend.DataResponse {
	bucket, err := d.reductClient.GetBucket(ctx, bucketName)
	if err != nil {
		log.DefaultLogger.Error("Failed to get bucket", "error", err)
		var apiErr *model.APIError
		errors.As(err, &apiErr)
		return backend.ErrDataResponse(backend.Status(apiErr.Status), apiErr.Message)
	}
	records, err := bucket.QueryMany(ctx, entries, &options)
	if err != nil {
		log.DefaultLogger.Error("Failed to query", "error", err)
		var apiErr *model.APIError
		errors.As(err, &apiErr)
		return backend.ErrDataResponse(backend.Status(apiErr.Status), apiErr.Message)
	}

	var frames []*data.Frame
	if combinedFrame {
		frames = getCombinedFrame(records.Records(), mode)
	} else {
		frames = getFrames(records.Records(), mode)
	}
	if err := records.Err(); err != nil {
		log.DefaultLogger.Error("Failed to stream records", "error", err)
		var apiErr *model.APIError
		errors.As(err, &apiErr)
		return backend.ErrDataResponse(backend.Status(apiErr.Status), apiErr.Message)
	}
	return backend.DataResponse{
		Frames: frames,
	}
}

type combinedRow struct {
	time   time.Time
	entry  string
	values map[string]any
}

const (
	labelColumnPrefix   = "label\x00"
	contentColumnPrefix = "content\x00"
)

// getCombinedFrame buffers records to discover a stable, nullable table schema.
func getCombinedFrame(records <-chan *reductgo.ReadableRecord, mode ReductMode) []*data.Frame {
	rows := make([]combinedRow, 0)
	labelKeys := make(map[string]struct{})
	contentKeys := make(map[string]struct{})

	for record := range records {
		row := combinedRow{time: time.UnixMicro(record.Time()), entry: record.Entry(), values: make(map[string]any)}
		if mode == "" || mode == ModeLabelOnly || mode == ModeLabelAndContent {
			for key, value := range record.Labels() {
				row.values[labelColumnPrefix+key] = fmt.Sprintf("%v", value)
				labelKeys[key] = struct{}{}
			}
		}
		if mode == ModeContentOnly || mode == ModeLabelAndContent {
			for key, value := range combinedContentValues(record) {
				row.values[contentColumnPrefix+key] = value
				contentKeys[key] = struct{}{}
			}
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil
	}

	columns := combinedColumnNames(labelKeys, contentKeys)
	fields := []*data.Field{
		data.NewField("time", nil, combinedTimes(rows)),
		data.NewField("entry", nil, combinedEntries(rows)),
	}
	for _, column := range columns {
		fields = append(fields, combinedField(column, rows))
	}
	frame := data.NewFrame("records", fields...)
	frame.Meta = &data.FrameMeta{Type: data.FrameTypeTable}
	return []*data.Frame{frame}
}

func combinedContentValues(record *reductgo.ReadableRecord) map[string]any {
	s, err := record.ReadAsString()
	if err != nil || len(strings.TrimSpace(s)) == 0 || !looksLikeJSON([]byte(s)) {
		return nil
	}
	var value any
	if json.Unmarshal([]byte(s), &value) != nil {
		return nil
	}
	flat := make(map[string]any)
	flattenJSON("$", value, flat)
	return flat
}

type combinedColumn struct {
	name string
	key  string
}

func combinedColumnNames(labelKeys, contentKeys map[string]struct{}) []combinedColumn {
	contentNames := make([]string, 0, len(contentKeys))
	used := map[string]struct{}{"time": {}, "entry": {}}
	for key := range contentKeys {
		contentNames = append(contentNames, key)
		used[key] = struct{}{}
	}
	sort.Strings(contentNames)

	columns := make([]combinedColumn, 0, len(labelKeys)+len(contentKeys))
	for _, key := range contentNames {
		columns = append(columns, combinedColumn{name: key, key: contentColumnPrefix + key})
	}
	labelNames := make([]string, 0, len(labelKeys))
	for key := range labelKeys {
		labelNames = append(labelNames, key)
	}
	sort.Strings(labelNames)
	for _, key := range labelNames {
		name := key
		if _, exists := used[name]; exists {
			name = "label." + key
		}
		base := name
		for suffix := 2; ; suffix++ {
			if _, exists := used[name]; !exists {
				break
			}
			name = fmt.Sprintf("%s.%d", base, suffix)
		}
		used[name] = struct{}{}
		columns = append(columns, combinedColumn{name: name, key: labelColumnPrefix + key})
	}
	sort.Slice(columns, func(i, j int) bool { return columns[i].name < columns[j].name })
	return columns
}

func combinedTimes(rows []combinedRow) []*time.Time {
	values := make([]*time.Time, len(rows))
	for i := range rows {
		values[i] = &rows[i].time
	}
	return values
}

func combinedEntries(rows []combinedRow) []*string {
	values := make([]*string, len(rows))
	for i := range rows {
		values[i] = &rows[i].entry
	}
	return values
}

func combinedField(column combinedColumn, rows []combinedRow) *data.Field {
	var kind reflect.Kind
	for _, row := range rows {
		value, exists := row.values[column.key]
		if !exists || value == nil {
			continue
		}
		if strings.HasPrefix(column.key, labelColumnPrefix) {
			value = parseValue(value.(string))
		}
		kind = reflect.TypeOf(value).Kind()
		break
	}

	switch kind {
	case reflect.Int64:
		values := make([]*int64, len(rows))
		for i, row := range rows {
			values[i] = combinedIntValue(row, column, kind)
		}
		return data.NewField(column.name, nil, values)
	case reflect.Float64:
		values := make([]*float64, len(rows))
		for i, row := range rows {
			values[i] = combinedFloatValue(row, column, kind)
		}
		return data.NewField(column.name, nil, values)
	case reflect.Bool:
		values := make([]*bool, len(rows))
		for i, row := range rows {
			values[i] = combinedBoolValue(row, column, kind)
		}
		return data.NewField(column.name, nil, values)
	default:
		values := make([]*string, len(rows))
		for i, row := range rows {
			values[i] = combinedStringValue(row, column, kind)
		}
		return data.NewField(column.name, nil, values)
	}
}

func combinedValue(row combinedRow, column combinedColumn, kind reflect.Kind) any {
	value, exists := row.values[column.key]
	if !exists || value == nil {
		return nil
	}
	if strings.HasPrefix(column.key, labelColumnPrefix) {
		parsed := parseValue(value.(string))
		if reflect.TypeOf(parsed).Kind() == kind {
			return parsed
		}
		coerced, err := coerceToKind(value.(string), kind)
		if err != nil {
			return nil
		}
		return coerced
	}
	if reflect.TypeOf(value).Kind() != kind {
		return nil
	}
	return value
}

func combinedIntValue(row combinedRow, column combinedColumn, kind reflect.Kind) *int64 {
	if value, ok := combinedValue(row, column, kind).(int64); ok {
		return &value
	}
	return nil
}
func combinedFloatValue(row combinedRow, column combinedColumn, kind reflect.Kind) *float64 {
	if value, ok := combinedValue(row, column, kind).(float64); ok {
		return &value
	}
	return nil
}
func combinedBoolValue(row combinedRow, column combinedColumn, kind reflect.Kind) *bool {
	if value, ok := combinedValue(row, column, kind).(bool); ok {
		return &value
	}
	return nil
}
func combinedStringValue(row combinedRow, column combinedColumn, kind reflect.Kind) *string {
	if value, ok := combinedValue(row, column, kind).(string); ok {
		return &value
	}
	return nil
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
	keys := make([]string, 0, len(frames))
	for k := range frames {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		result = append(result, frames[k])
	}
	return result
}

// processLabels processes the labels of a record and appends them to the frames.
func processLabels(frames map[string]*data.Frame, kindMap map[string]reflect.Kind, record *reductgo.ReadableRecord) {
	entryName := record.Entry()
	for key, labelValue := range record.Labels() {
		frameKey := entryName + "/" + key

		strValue := fmt.Sprintf("%v", labelValue)
		initialType, ok := kindMap[frameKey]
		value := parseValue(strValue)

		if !ok {
			kind := reflect.TypeOf(value).Kind()
			kindMap[frameKey] = kind
			initialType = kind
		}

		currentType := reflect.TypeOf(value).Kind()
		if currentType != initialType {
			// If the type has changed, we need to coerce the value to the initial type
			log.DefaultLogger.Debug("Type change detected", "key", frameKey, "from", initialType, "to", currentType)
			val, err := coerceToKind(strValue, initialType)
			if err != nil {
				log.DefaultLogger.Error("Failed to coerce value", "key", frameKey, "value", strValue, "error", err)
				continue
			}
			value = val
		}

		switch v := value.(type) {
		case int64:
			appendValue(frames, frameKey, record, v)
		case float64:
			appendValue(frames, frameKey, record, v)
		case bool:
			appendValue(frames, frameKey, record, v)
		case string:
			appendValue(frames, frameKey, record, v)
		default:
			appendValue(frames, frameKey, record, strValue)
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

	entryName := record.Entry()
	for k, val := range flat {
		// Create entry-prefixed frame key to separate time series per entry
		frameKey := entryName + "/" + k
		switch v := val.(type) {
		case int64:
			appendValue(frames, frameKey, record, v)
		case float64:
			appendValue(frames, frameKey, record, v)
		case bool:
			appendValue(frames, frameKey, record, v)
		case string:
			appendValue(frames, frameKey, record, v)
		default:
			str := fmt.Sprintf("%v", val)
			appendValue(frames, frameKey, record, str)
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
