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
		reductgo.NewReadableRecord("entry-0", time.Now().UnixMicro(), 0, true, io.NopCloser(strings.NewReader("")), reductgo.LabelMap{
			"intLabel":    "42",
			"floatLabel":  "3.14",
			"boolLabel":   "true",
			"stringLabel": "hello",
		}, ""),
		reductgo.NewReadableRecord("entry-1", time.Now().Add(time.Second).UnixMicro(), 0, true, io.NopCloser(strings.NewReader("")), reductgo.LabelMap{
			"intLabel":    "21.9", // should truncate to 21
			"floatLabel":  "6",
			"boolLabel":   "false", // should become false
			"stringLabel": "world",
		}, ""),
		reductgo.NewReadableRecord("entry-2", time.Now().Add(2*time.Second).UnixMicro(), 0, true, io.NopCloser(strings.NewReader("")), reductgo.LabelMap{
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
	assert.Equal(t, data.NewField("value", nil, []float64{3.14, 6.0}), floatFrame.Fields[1])

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

func TestProcessContent_PreservesJSONTypes(t *testing.T) {
	frames := make(map[string]*data.Frame)

	jsonContent1 := `{
		"str_number": "123",
		"source_id": "00000001_000",
		"temp": 25.5,
		"flag": true,
		"count": 42
	}`

	record1 := reductgo.NewReadableRecord(
		"json-entry-0",
		time.Now().UnixMicro(),
		0,
		true,
		io.NopCloser(strings.NewReader(jsonContent1)),
		reductgo.LabelMap{},
		"application/json",
	)

	jsonContent2 := `{
		"str_number": "456",
		"source_id": "00000002_001",
		"temp": 30.0,
		"flag": false,
		"count": 84
	}`

	record2 := reductgo.NewReadableRecord(
		"json-entry-1",
		time.Now().Add(time.Second).UnixMicro(),
		0,
		true,
		io.NopCloser(strings.NewReader(jsonContent2)),
		reductgo.LabelMap{},
		"application/json",
	)

	processContent(frames, record1)
	processContent(frames, record2)

	strNumFrame, exists := frames["$.str_number"]
	assert.True(t, exists, "$.str_number frame should exist")
	assert.Equal(t, data.FieldTypeString, strNumFrame.Fields[1].Type())
	assert.Equal(t, "123", strNumFrame.Fields[1].At(0))
	assert.Equal(t, "456", strNumFrame.Fields[1].At(1))

	sourceIdFrame, exists := frames["$.source_id"]
	assert.True(t, exists, "$.source_id frame should exist")
	assert.Equal(t, data.FieldTypeString, sourceIdFrame.Fields[1].Type())
	assert.Equal(t, "00000001_000", sourceIdFrame.Fields[1].At(0))
	assert.Equal(t, "00000002_001", sourceIdFrame.Fields[1].At(1))

	tempFrame, exists := frames["$.temp"]
	assert.True(t, exists, "$.temp frame should exist")
	assert.Equal(t, data.FieldTypeFloat64, tempFrame.Fields[1].Type())
	assert.Equal(t, 25.5, tempFrame.Fields[1].At(0))
	assert.Equal(t, 30.0, tempFrame.Fields[1].At(1))

	flagFrame, exists := frames["$.flag"]
	assert.True(t, exists, "$.flag frame should exist")
	assert.Equal(t, data.FieldTypeBool, flagFrame.Fields[1].Type())
	assert.Equal(t, true, flagFrame.Fields[1].At(0))
	assert.Equal(t, false, flagFrame.Fields[1].At(1))

	countFrame, exists := frames["$.count"]
	assert.True(t, exists, "$.count frame should exist")
	assert.Equal(t, data.FieldTypeFloat64, countFrame.Fields[1].Type())
	assert.Equal(t, float64(42), countFrame.Fields[1].At(0))
	assert.Equal(t, float64(84), countFrame.Fields[1].At(1))
}
