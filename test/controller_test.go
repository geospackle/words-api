package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	controller "github.com/geospackle/go-words-api/src/api"
	"github.com/geospackle/go-words-api/src/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOpenSearchRepository struct {
	mock.Mock
}

func (m *mockOpenSearchRepository) Insert(ctx context.Context, index string, document string) error {
	args := m.Called(ctx, index, document)
	return args.Error(0)
}

func (m *mockOpenSearchRepository) Search(ctx context.Context, index []string, query string) (*repository.SearchResult, error) {
	args := m.Called(ctx, index, query)
	return args.Get(0).(*repository.SearchResult), args.Error(1)
}

func TestPostHandler_Success(t *testing.T) {
	mockRepo := new(mockOpenSearchRepository)
	mockRepo.On("Insert", context.Background(), "test-index", mock.AnythingOfType("string")).Return(nil)
	payload := controller.Payload{Word: "test"}
	reqBody, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))

	w := httptest.NewRecorder()

	controller.PostHandler(w, req, "test-index", mockRepo)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	expected := fmt.Sprintf("Successfully inserted value: %s", "test")
	assert.Equal(t, expected, string(body))

	mockRepo.AssertNumberOfCalls(t, "Insert", 1)
	mockRepo.AssertExpectations(t)
}

func TestPostHandler_MethodNotAllowed(t *testing.T) {
	mockRepo := &mockOpenSearchRepository{}

	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	w := httptest.NewRecorder()

	controller.PostHandler(w, req, "test-index", mockRepo)

	assert.Contains(t, w.Body.String(), "Method not allowed")
	assert.Equal(t, w.Code, http.StatusMethodNotAllowed)
}

func TestPostHandler_InvalidRequestBody(t *testing.T) {
	mockRepo := &mockOpenSearchRepository{}

	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("{invalid_json}")))

	w := httptest.NewRecorder()

	controller.PostHandler(w, req, "test-index", mockRepo)

	assert.Contains(t, w.Body.String(), "Invalid request body")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}

func TestPostHandler_InvalidPayload(t *testing.T) {
	mockRepo := &mockOpenSearchRepository{}

	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"not_word": "value"}`)))

	w := httptest.NewRecorder()

	controller.PostHandler(w, req, "test-index", mockRepo)

	assert.Contains(t, w.Body.String(), "Invalid input payload")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}

func TestPostHandler_EmptyPayload(t *testing.T) {
	mockRepo := &mockOpenSearchRepository{}

	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(``)))

	w := httptest.NewRecorder()

	controller.PostHandler(w, req, "test-index", mockRepo)

	assert.Contains(t, w.Body.String(), "EOF")
	assert.Equal(t, w.Code, http.StatusBadRequest)
}

func TestPostHandler_InsertError(t *testing.T) {
	mockRepo := new(mockOpenSearchRepository)
	mockRepo.On("Insert", context.Background(), "test-index", mock.AnythingOfType("string")).Return(errors.New("insert error"))

	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"word": "test"}`)))

	w := httptest.NewRecorder()

	controller.PostHandler(w, req, "test-index", mockRepo)

	assert.Contains(t, w.Body.String(), "Error inserting value")
	assert.Equal(t, w.Code, http.StatusBadGateway)
	mockRepo.AssertNumberOfCalls(t, "Insert", 1)
	mockRepo.AssertExpectations(t)

}

func TestGetHandler_Success(t *testing.T) {
	mockRepo := new(mockOpenSearchRepository)
	searchResult := repository.SearchResult{
		Aggregations: repository.Aggregations{
			MaxDistinctCounts: struct {
				Keys  []string `json:"keys"`
				Value float32  `json:"value"`
			}{
				Keys:  []string{"key1"},
				Value: 10.0,
			},
		},
	}

	mockRepo.On("Search", context.Background(), mock.Anything, mock.Anything).Return(&searchResult, nil)

	req, _ := http.NewRequest(http.MethodGet, "/?prefix=test", nil)

	w := httptest.NewRecorder()

	controller.GetHandler(w, req, []string{"my-index"}, mockRepo)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	fmt.Println(w.Body)
	var response controller.Response
	_ = json.NewDecoder(w.Body).Decode(&response)

	expectedData := map[string]interface{}{"keys": []interface{}{"key1"}, "value": 10.0}
	assert.Equal(t, response.Data, expectedData)

	mockRepo.AssertExpectations(t)
}

func TestGetHandler_MissingQueryParameter(t *testing.T) {
	mockRepo := new(mockOpenSearchRepository)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	w := httptest.NewRecorder()

	controller.GetHandler(w, req, []string{"my-index"}, mockRepo)

	assert.Equal(t, w.Code, http.StatusBadRequest)
	expectedBody := "Needs query parameter 'word'\n"
	responseBody := w.Body.String()
	assert.Equal(t, expectedBody, responseBody)
}

func TestGetHandler_RepositoryError(t *testing.T) {
	mockRepo := new(mockOpenSearchRepository)
	expectedErr := fmt.Errorf("search error")
	mockRepo.On("Search", context.Background(), mock.Anything, mock.Anything).Return(&repository.SearchResult{}, expectedErr)

	req, _ := http.NewRequest(http.MethodGet, "/?prefix=test", nil)

	w := httptest.NewRecorder()

	controller.GetHandler(w, req, []string{"my-index"}, mockRepo)

	assert.Equal(t, w.Code, http.StatusBadGateway)

	expectedBody := fmt.Sprintf("Query can not be processed: %s\n", expectedErr)
	responseBody := w.Body.String()
	assert.Equal(t, responseBody, expectedBody)
}

func TestValidateValue(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		expect bool
	}{
		{name: "valid alphabetic string", value: "Hello", expect: true},
		{name: "empty string", value: "", expect: false},
		{name: "string with numbers", value: "Hello123", expect: false},
		{name: "string with special characters", value: "Hello!", expect: false},
		{name: "empty value", value: "", expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := controller.ValidateValue(tc.value)
			assert.Equal(t, tc.expect, got)
		})
	}
}
