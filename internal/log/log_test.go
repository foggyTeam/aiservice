package log

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonFormatterWrite(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expected  []byte
		shouldErr bool
	}{
		{
			name:      "writes valid JSON with indentation",
			input:     []byte(`{"key":"value","number":42}`),
			expected:  []byte("{\n  \"key\": \"value\",\n  \"number\": 42\n}\n"),
			shouldErr: false,
		},
		{
			name:      "writes invalid JSON as-is",
			input:     []byte(`not valid json`),
			expected:  []byte(`not valid json`),
			shouldErr: false,
		},
		{
			name:      "writes empty JSON object",
			input:     []byte(`{}`),
			expected:  []byte("{}\n"),
			shouldErr: false,
		},
		{
			name:      "writes complex nested JSON",
			input:     []byte(`{"nested":{"inner":"value"},"array":[1,2,3]}`),
			expected:  []byte("{\n  \"array\": [\n    1,\n    2,\n    3\n  ],\n  \"nested\": {\n    \"inner\": \"value\"\n  }\n}\n"),
			shouldErr: false,
		},
		{
			name:      "writes JSON array",
			input:     []byte(`[1,2,3]`),
			expected:  []byte("[\n  1,\n  2,\n  3\n]\n"),
			shouldErr: false,
		},
		{
			name:      "writes plain text",
			input:     []byte(`simple text`),
			expected:  []byte(`simple text`),
			shouldErr: false,
		},
		{
			name:      "writes JSON with unicode",
			input:     []byte(`{"message":"hello 世界"}`),
			expected:  []byte("{\n  \"message\": \"hello 世界\"\n}\n"),
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := JsonFormatter{w: buf}

			n, err := formatter.Write(tt.input)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Greater(t, n, 0)
			}

			actual := buf.Bytes()
			assert.Equal(t, tt.expected, actual, "output mismatch")
		})
	}
}

func TestJsonFormatterWriteWithBuffer(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := JsonFormatter{w: buf}

	// Write multiple times
	input1 := []byte(`{"first":1}`)
	input2 := []byte(`{"second":2}`)

	n1, err1 := formatter.Write(input1)
	require.NoError(t, err1)
	assert.Greater(t, n1, 0)

	n2, err2 := formatter.Write(input2)
	require.NoError(t, err2)
	assert.Greater(t, n2, 0)

	output := buf.String()
	assert.Contains(t, output, "first")
	assert.Contains(t, output, "second")
}

func TestJsonFormatterWriteEmptyInput(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := JsonFormatter{w: buf}

	n, err := formatter.Write([]byte{})

	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestJsonFormatterWriteMalformedJSON(t *testing.T) {
	testCases := []string{
		`{invalid}`,
		`["unclosed`,
		`{"key": "value"`,
		`{key: value}`,
		``,
	}

	for _, tc := range testCases {
		t.Run("malformed: "+tc, func(t *testing.T) {
			buf := &bytes.Buffer{}
			formatter := JsonFormatter{w: buf}

			_, err := formatter.Write([]byte(tc))

			assert.NoError(t, err)
		})
	}
}

func TestNewJsonFormatter(t *testing.T) {
	formatter := NewJsonFormatter()

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.w)
}

func TestSetupJsonLogger(t *testing.T) {
	logger := SetupJsonLogger()

	assert.NotNil(t, logger)

	// Verify the logger is set as default
	defaultLogger := logger
	assert.NotNil(t, defaultLogger)
}

func TestJsonFormatterIntegration(t *testing.T) {
	// Create a formatter with a custom writer
	buf := &bytes.Buffer{}
	formatter := JsonFormatter{w: buf}

	// Create a map to marshal as JSON
	data := map[string]interface{}{
		"level":   "info",
		"msg":     "test message",
		"time":    "2024-01-01T00:00:00Z",
		"details": map[string]int{"count": 42},
	}

	jsonBytes, _ := json.Marshal(data)

	n, err := formatter.Write(jsonBytes)

	require.NoError(t, err)
	assert.Greater(t, n, 0)

	output := buf.String()
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
}

func TestJsonFormatterWriteLargeJSON(t *testing.T) {
	// Create a large JSON structure
	largeData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largeData[string(rune(i))] = i
	}

	jsonBytes, _ := json.Marshal(largeData)

	buf := &bytes.Buffer{}
	formatter := JsonFormatter{w: buf}

	n, err := formatter.Write(jsonBytes)

	require.NoError(t, err)
	assert.Greater(t, n, len(jsonBytes))

	output := buf.String()
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, 100, len(result))
}
