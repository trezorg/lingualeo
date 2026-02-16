package api

import (
	"encoding/json/v2"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertibleBooleanUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    bool
		wantErr bool
	}{
		{
			name:  "string true",
			input: `"true"`,
			want:  true,
		},
		{
			name:  "string false",
			input: `"false"`,
			want:  false,
		},
		{
			name:  "string 1",
			input: `"1"`,
			want:  true,
		},
		{
			name:  "string 0",
			input: `"0"`,
			want:  false,
		},
		{
			name:  "null",
			input: `null`,
			want:  false,
		},
		{
			name:  "boolean true",
			input: `true`,
			want:  true,
		},
		{
			name:  "boolean false",
			input: `false`,
			want:  false,
		},
		{
			name:    "invalid string",
			input:   `"yes"`,
			wantErr: true,
		},
		{
			name:    "invalid number",
			input:   `2`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got convertibleBoolean
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, errBooleanUnmarshal))
			} else {
				require.NoError(t, err)
				assert.Equal(t, convertibleBoolean(tt.want), got)
			}
		})
	}
}

func TestCheckAuthError(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name: "empty body",
			body: "",
		},
		{
			name: "success response",
			body: `{"error_msg": "", "error_code": 0}`,
		},
		{
			name:    "error response",
			body:    `{"error_msg": "Invalid credentials", "error_code": 401}`,
			wantErr: true,
		},
		{
			name: "no error code",
			body: `{"error_msg": "", "error_code": 0}`,
		},
		{
			name:    "invalid json",
			body:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkAuthError([]byte(tt.body))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResultFromResponse(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantErr      bool
		wantWord     string
		wantErrorMsg string
	}{
		{
			name: "valid translation response",
			body: `{
				"sound_url": "https://example.com/sound.mp3",
				"transcription": "test",
				"translate": [
					{"value": "тест", "votes": 10, "id": 1},
					{"value": "проверка", "votes": 5, "id": 2}
				]
			}`,
			wantWord: "test",
		},
		{
			name: "response with error",
			body: `{
				"error_msg": "Word not found",
				"sound_url": "",
				"transcription": "",
				"translate": []
			}`,
			wantErr:      true,
			wantErrorMsg: "Word not found",
		},
		{
			name:    "invalid json",
			body:    `{invalid}`,
			wantErr: true,
		},
		{
			name:     "no result fallback",
			body:     `{"error_msg": "API error", "translate": ["test1", "test2"]}`,
			wantErr:  true,
			wantWord: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Word: tt.wantWord}
			err := r.FromResponse([]byte(tt.body))
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrorMsg != "" {
					var resultErr ResultError
					require.True(t, errors.As(err, &resultErr))
					assert.Contains(t, resultErr.Error(), tt.wantErrorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResultInDictionary(t *testing.T) {
	tests := []struct {
		name      string
		result    Result
		wantExist bool
	}{
		{
			name: "exists in dictionary - user flag",
			result: Result{
				Exists: true,
			},
			wantExist: true,
		},
		{
			name: "exists in dictionary - word flag",
			result: Result{
				Exists: false,
				Translate: []Word{
					{Value: "test", Exists: true},
				},
			},
			wantExist: true,
		},
		{
			name: "not in dictionary",
			result: Result{
				Exists: false,
				Translate: []Word{
					{Value: "test", Exists: false},
				},
			},
			wantExist: false,
		},
		{
			name: "empty translate",
			result: Result{
				Exists:    false,
				Translate: []Word{},
			},
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.InDictionary()
			assert.Equal(t, tt.wantExist, got)
		})
	}
}

func TestResultIsRussian(t *testing.T) {
	tests := []struct {
		name          string
		transcription string
		wantIsRussian bool
	}{
		{
			name:          "Russian word - no transcription",
			transcription: "",
			wantIsRussian: true,
		},
		{
			name:          "English word - has transcription",
			transcription: "ˈtest",
			wantIsRussian: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Transcription: tt.transcription}
			got := r.IsRussian()
			assert.Equal(t, tt.wantIsRussian, got)
		})
	}
}

func TestResultError(t *testing.T) {
	tests := []struct {
		name    string
		result  Result
		wantMsg string
	}{
		{
			name:    "with error message",
			result:  Result{Word: "test", ErrorMsg: "Word not found"},
			wantMsg: "test: Word not found",
		},
		{
			name:    "without error message",
			result:  Result{Word: "test", ErrorMsg: ""},
			wantMsg: "cannot translate word",
		},
		{
			name:    "error message only - no word",
			result:  Result{ErrorMsg: "API error"},
			wantMsg: "API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ResultError{Result: tt.result}
			assert.Equal(t, tt.wantMsg, err.Error())
			assert.True(t, errors.Is(err, errTranslateWord))
		})
	}
}

func TestResultSetTranslation(t *testing.T) {
	r := Result{}
	r.SetTranslation([]string{"test1", "test2", "test1", "test3"})
	assert.Equal(t, []string{"test1", "test2", "test3"}, r.AddWords)
}

func TestResultHasError(t *testing.T) {
	tests := []struct {
		name      string
		errorMsg  string
		wantError bool
	}{
		{
			name:      "has error",
			errorMsg:  "Something went wrong",
			wantError: true,
		},
		{
			name:      "no error",
			errorMsg:  "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{ErrorMsg: tt.errorMsg}
			assert.Equal(t, tt.wantError, r.HasError())
		})
	}
}

func TestResultParse(t *testing.T) {
	r := Result{
		Translate: []Word{
			{Value: "test", Votes: 5},
			{Value: "test", Votes: 3}, // duplicate
			{Value: "check", Votes: 10},
		},
	}
	r.parse()

	// Should be unique and sorted by votes
	require.Len(t, r.Translate, 2)
	assert.Equal(t, "check", r.Translate[0].Value) // 10 votes
	assert.Equal(t, "test", r.Translate[1].Value)  // 5 votes
}

func TestOpResultFromBody(t *testing.T) {
	body := `{"sound_url": "https://example.com/sound.mp3", "transcription": "test"}`
	result := opResultFromBody("testword", []byte(body))
	assert.Equal(t, "testword", result.Result.Word)
	assert.NoError(t, result.Error)
}
