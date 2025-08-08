package plugin

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	reductgo "github.com/reductstore/reduct-go"
	"github.com/stretchr/testify/assert"
)

func TestProcessLabels(t *testing.T) {
	frames := make(map[string]*data.Frame)

	records := []*reductgo.ReadableRecord{
		reductgo.NewReadableRecord(time.Now().UnixMicro(), 0, true, io.NopCloser(strings.NewReader("")), reductgo.LabelMap{
			"intLabel":    "42",
			"floatLabel":  "3.14",
			"boolLabel":   "true",
			"stringLabel": "hello",
		}, ""),
		reductgo.NewReadableRecord(time.Now().Add(time.Second).UnixMicro(), 0, true, io.NopCloser(strings.NewReader("")), reductgo.LabelMap{
			"intLabel":    "21.9", // should truncate to 21
			"floatLabel":  "6",
			"boolLabel":   "false", // should become false
			"stringLabel": "world",
		}, ""),
		reductgo.NewReadableRecord(time.Now().Add(2*time.Second).UnixMicro(), 0, true, io.NopCloser(strings.NewReader("")), reductgo.LabelMap{
			"intLabel":    "badInt",   // fallback should be ignored
			"floatLabel":  "badFloat", // fallback should be ignored
			"boolLabel":   "notBool",  // fallback should be ignored
			"stringLabel": "stay",
		}, ""),
	}

	labelInitialType := make(map[string]reflect.Kind)
	for _, rec := range records {
		processLabels(frames, labelInitialType, rec)
	}

	assert.Len(t, frames, 4)

	// Test intLabel frame
	intFrame := frames["intLabel"]
	assert.Equal(t, 2, len(intFrame.Fields))
	assert.Equal(t, data.FieldTypeInt64, intFrame.Fields[1].Type())
	assert.Equal(t, data.NewField("value", nil, []int64{42, 21}), intFrame.Fields[1])

	// Test floatLabel frame
	floatFrame := frames["floatLabel"]
	assert.Equal(t, 2, len(floatFrame.Fields))
	assert.Equal(t, data.FieldTypeFloat64, floatFrame.Fields[1].Type())
	assert.Equal(t, data.NewField("value", nil, []float64{3.14, 6.}), floatFrame.Fields[1])

	// Test boolLabel frame
	boolFrame := frames["boolLabel"]
	assert.Equal(t, 2, len(boolFrame.Fields))
	assert.Equal(t, data.FieldTypeBool, boolFrame.Fields[1].Type())
	assert.Equal(t, data.NewField("value", nil, []bool{true, false}), boolFrame.Fields[1])

	// Test stringLabel frame
	strFrame := frames["stringLabel"]
	assert.Equal(t, 2, len(strFrame.Fields))
	assert.Equal(t, data.FieldTypeString, strFrame.Fields[1].Type())
	assert.Equal(t, data.NewField("value", nil, []string{"hello", "world", "stay"}), strFrame.Fields[1])
}
