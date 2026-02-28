package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		body       []byte
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful GET request",
			method:     http.MethodGet,
			response:   `{"status": "ok"}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "successful POST request with body",
			method:     http.MethodPost,
			body:       []byte(`{"test": "data"}`),
			response:   `{"status": "ok"}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "error status code",
			method:     http.MethodGet,
			response:   `{"error": "not found"}`,
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "internal server error",
			method:     http.MethodGet,
			response:   `{"error": "internal error"}`,
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "application/json", r.Header.Get("Accept"))

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			ctx := t.Context()
			api := &API{
				client:  http.DefaultClient,
				Debug:   false,
				timeout: 10 * time.Second,
				retryConfig: RetryConfig{
					MaxAttempts: 1, // Only one attempt for tests
				},
			}
			resp, err := api.request(ctx, requestParams{
				method: tt.method,
				url:    server.URL,
				body:   tt.body,
			})
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "unexpected response status code")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.response, string(resp))
			}
		})
	}
}

func TestRequestWithContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := t.Context()
	api := &API{
		client:  http.DefaultClient,
		Debug:   false,
		timeout: 10 * time.Millisecond,
		retryConfig: RetryConfig{
			MaxAttempts: 1, // Only one attempt for tests
		},
	}

	_, err := api.request(ctx, requestParams{
		method: http.MethodGet,
		url:    server.URL,
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, errAPIRequestTimeout))
}

func TestRequestWithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameter is set
		assert.Equal(t, "word=test", r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	ctx := t.Context()
	api := &API{
		client:  http.DefaultClient,
		Debug:   false,
		timeout: 10 * time.Second,
		retryConfig: RetryConfig{
			MaxAttempts: 1, // Only one attempt for tests
		},
	}
	resp, err := api.request(ctx, requestParams{
		method: http.MethodGet,
		url:    server.URL,
		query:  "word=test",
	})
	require.NoError(t, err)
	assert.Equal(t, `{"status": "ok"}`, string(resp))
}

func TestPrepareClient(t *testing.T) {
	cfg := DefaultConfig()
	client, err := prepareClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.Jar)
	assert.NotNil(t, client.Transport)
}

func TestAPIStructFields(t *testing.T) {
	api := &API{
		Email:    "test@example.com",
		Password: "password123",
		Debug:    true,
	}

	assert.Equal(t, "test@example.com", api.Email)
	assert.Equal(t, "password123", api.Password)
	assert.True(t, api.Debug)
}

func TestOpResultFromBodyWithError(t *testing.T) {
	// Test with error response
	body := `{"error_msg": "Translation failed", "translate": []}`
	result := opResultFromBody("testword", []byte(body))
	assert.Equal(t, "testword", result.Result.Word)
	assert.Error(t, result.Error)
}

func TestClientInterface(t *testing.T) {
	// Verify API implements Client interface
	var _ Client = (*API)(nil)
}

// MockClient implements Client for testing higher-level code
type MockClient struct {
	TranslateResult OperationResult
	AddResult       OperationResult
}

func (m *MockClient) TranslateWord(_ context.Context, _ string) OperationResult {
	return m.TranslateResult
}

func (m *MockClient) AddWord(_ context.Context, _, _ string) OperationResult {
	return m.AddResult
}

func TestMockClientImplementsInterface(t *testing.T) {
	var _ Client = &MockClient{}
}

func TestMockClientTranslate(t *testing.T) {
	mock := &MockClient{
		TranslateResult: OperationResult{
			Result: Result{
				Word:          "hello",
				Transcription: "həˈləʊ",
				Translate: []Word{
					{Value: "привет", Votes: 100},
				},
			},
		},
	}

	result := mock.TranslateWord(t.Context(), "hello")
	require.NoError(t, result.Error)
	assert.Equal(t, "hello", result.Result.Word)
	assert.Len(t, result.Result.Translate, 1)
}

func TestMockClientAddWord(t *testing.T) {
	mock := &MockClient{
		AddResult: OperationResult{
			Error: nil,
		},
	}

	result := mock.AddWord(t.Context(), "hello", "привет")
	require.NoError(t, result.Error)
}
