package digitalink

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aiservice/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func saveAndSetEndpoint(newEndpoint string) func() {
	original := googleHandwritingEndpoint
	googleHandwritingEndpoint = newEndpoint
	return func() {
		googleHandwritingEndpoint = original
	}
}

func TestTransformToGoogleTrace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		elements []models.Element
		wantLen  int
	}{
		{
			name: "элементы с точками",
			elements: []models.Element{
				{Type: "line", Points: []float32{10, 20, 30, 40, 50, 60}},
			},
			wantLen: 1,
		},
		{
			name: "элементы без точек",
			elements: []models.Element{
				{Type: "line", X: 100, Y: 200, Width: 50, Height: 30},
			},
			wantLen: 1,
		},
		{
			name: "игнорирование не-line элементов",
			elements: []models.Element{
				{Type: "text", Content: "hello"},
				{Type: "rectangle"},
				{Type: "line", Points: []float32{1, 2, 3, 4}},
			},
			wantLen: 1,
		},
		{
			name: "несколько штрихов",
			elements: []models.Element{
				{Type: "line", Points: []float32{0, 0, 10, 10}},
				{Type: "text"},
				{Type: "line", Points: []float32{20, 20, 30, 30}},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := transformToGoogleTrace(tt.elements)
			assert.Len(t, got, tt.wantLen)

			for _, stroke := range got {
				require.Len(t, stroke, 2, "[xCoords, yCoords]")
				assert.IsType(t, []any{}, stroke[0], "xCoords должен быть []any")
				assert.IsType(t, []any{}, stroke[1], "yCoords должен быть []any")
			}
		})
	}
}

func TestFindMaxDimensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		elements []models.Element
		wantX    float32
		wantY    float32
	}{
		{
			name: "один элемент",
			elements: []models.Element{
				{X: 10, Y: 20, Width: 100, Height: 50},
			},
			wantX: 110,
			wantY: 70,
		},
		{
			name:     "пустой список - 0,0",
			elements: []models.Element{},
			wantX:    0,
			wantY:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotX, gotY := findMaxDimensions(tt.elements)
			assert.InDelta(t, tt.wantX, gotX, 0.001)
			assert.InDelta(t, tt.wantY, gotY, 0.001)
		})
	}
}

func TestExtractRecognizedTextFromRaw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawResp  RawGoogleHandwritingResponse
		expected string
	}{
		{
			name: "успешный ответ",
			rawResp: RawGoogleHandwritingResponse{
				"SUCCESS",
				[]any{
					[]any{
						"candidate-id",
						[]any{"распознанный текст", "альтернатива"},
						[]any{},
						map[string]any{},
					},
				},
				"extra_data",
			},
			expected: "распознанный текст",
		},
		{
			name: "",
			rawResp: RawGoogleHandwritingResponse{
				"SUCCESS",
				[]any{
					[]any{"id1", []any{"строка1"}, []any{}, map[string]any{}},
					[]any{"id2", []any{"строка2"}, []any{}, map[string]any{}},
				},
			},
			expected: "строка1\nстрока2",
		},
		{
			name: "пустой массив кандидатов",
			rawResp: RawGoogleHandwritingResponse{
				"SUCCESS",
				[]any{},
			},
			expected: "",
		},
		{
			name:     "короткий ответ",
			rawResp:  RawGoogleHandwritingResponse{"ERROR"},
			expected: "",
		},
		{
			name: "некорректный тип",
			rawResp: RawGoogleHandwritingResponse{
				"SUCCESS",
				"not-an-array",
			},
			expected: "",
		},
		{
			name: "пустой текст ",
			rawResp: RawGoogleHandwritingResponse{
				"SUCCESS",
				[]any{
					[]any{"id", []any{""}, []any{}, map[string]any{}},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractRecognizedTextFromRaw(tt.rawResp)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRecognizeInk_Success(t *testing.T) {
	restore := saveAndSetEndpoint("")
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("User-Agent"), "Mozilla/5.0")

		var req GoogleHandwritingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "enable_pre_space", req.Options)
		require.Len(t, req.Requests, 1)
		assert.Equal(t, "ru", req.Requests[0].Language)

		rawResponse := []any{
			"SUCCESS",
			[]any{
				[]any{"cand-1", []any{"Привет мир"}, []any{}, map[string]any{}},
			},
			map[string]any{},
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(rawResponse)
		require.NoError(t, err, "не удалось закодировать ответ")
	}))
	defer server.Close()

	restore = saveAndSetEndpoint(server.URL)
	defer restore()

	client := NewClient("ru", 5*time.Second)
	elements := []models.Element{
		{Type: "line", Points: []float32{10, 20, 30, 40}},
	}

	ctx := context.Background()
	text, err := client.RecognizeInk(ctx, elements)

	require.NoError(t, err)
	assert.Equal(t, "Привет мир", text)
}

func TestRecognizeInk_HTTPError(t *testing.T) {
	restore := saveAndSetEndpoint("")
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "service unavailable"}`))
	}))
	defer server.Close()

	restore = saveAndSetEndpoint(server.URL)
	defer restore()

	client := NewClient("en", 2*time.Second)

	_, err := client.RecognizeInk(context.Background(), []models.Element{
		{Type: "line", Points: []float32{1, 2}},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestRecognizeInk_InvalidJSONResponse(t *testing.T) {
	restore := saveAndSetEndpoint("")
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json {{{`))
	}))
	defer server.Close()

	restore = saveAndSetEndpoint(server.URL)
	defer restore()

	client := NewClient("en", 2*time.Second)

	_, err := client.RecognizeInk(context.Background(), []models.Element{
		{Type: "line"},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestRecognizeInk_ContextCancellation(t *testing.T) {
	restore := saveAndSetEndpoint("")
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer server.Close()

	restore = saveAndSetEndpoint(server.URL)
	defer restore()

	client := NewClient("en", 5*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.RecognizeInk(ctx, []models.Element{{Type: "line"}})

	require.Error(t, err)
}

func TestRecognizeInk_EmptyElements(t *testing.T) {
	restore := saveAndSetEndpoint("")
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GoogleHandwritingRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		assert.NotEmpty(t, req.Requests)
		if len(req.Requests) > 0 {
			assert.Len(t, req.Requests[0].Ink, 1)
		}

		_ = json.NewEncoder(w).Encode([]any{"SUCCESS", []any{}, nil})
	}))
	defer server.Close()

	restore = saveAndSetEndpoint(server.URL)
	defer restore()

	client := NewClient("en", 2*time.Second)

	text, err := client.RecognizeInk(context.Background(), []models.Element{})

	require.NoError(t, err)
	assert.Empty(t, text)
}

func TestRecognizeInk_WritingGuide(t *testing.T) {
	restore := saveAndSetEndpoint("")
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GoogleHandwritingRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if len(req.Requests) > 0 {
			guide := req.Requests[0].WritingGuide
			assert.Greater(t, guide.Width, 0)
			assert.Greater(t, guide.Height, 0)
		}

		_ = json.NewEncoder(w).Encode([]any{"SUCCESS", []any{}, nil})
	}))
	defer server.Close()

	restore = saveAndSetEndpoint(server.URL)
	defer restore()

	client := NewClient("en", 2*time.Second)

	elements := []models.Element{
		{Type: "line", X: 10, Y: 20, Width: 100, Height: 50},
	}
	_, _ = client.RecognizeInk(context.Background(), elements)
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := NewClient("ja", 10*time.Second)

	assert.NotNil(t, client.httpClient)
	assert.Equal(t, "ja", client.language)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestRecognizeInk_FullFlow(t *testing.T) {
	restore := saveAndSetEndpoint("")
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := []any{
			"SUCCESS",
			[]any{
				[]any{"c1", []any{"function main()"}, []any{}, map[string]any{}},
				[]any{"c2", []any{"  return 42;"}, []any{}, map[string]any{}},
			},
			nil,
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err, "")
	}))
	defer server.Close()

	restore = saveAndSetEndpoint(server.URL)
	defer restore()

	client := NewClient("en", 5*time.Second)

	elements := []models.Element{
		{Type: "line", Points: []float32{0, 0, 10, 10}},
		{Type: "line", Points: []float32{20, 20, 30, 30}},
		{Type: "text", Content: "ignored"},
	}

	text, err := client.RecognizeInk(context.Background(), elements)

	require.NoError(t, err)
	assert.Equal(t, "function main()\n  return 42;", text)
}
