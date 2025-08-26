package responses

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/supchaser/wb_l0/internal/utils/errs"
	"github.com/supchaser/wb_l0/internal/utils/logger"
)

func TestMain(m *testing.M) {
	logger.InitTestLogger()
	m.Run()
}

func TestDoBadResponseAndLog(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
	}{
		{
			name:       "BadRequest",
			statusCode: http.StatusBadRequest,
			message:    "Invalid input",
		},
		{
			name:       "NotFound",
			statusCode: http.StatusNotFound,
			message:    "Resource not found",
		},
		{
			name:       "InternalServerError",
			statusCode: http.StatusInternalServerError,
			message:    "Something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			DoBadResponseAndLog(w, tt.statusCode, tt.message)

			assert.Equal(t, tt.statusCode, w.Code)

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response BadResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.statusCode, response.Status)
			assert.Equal(t, tt.message, response.Text)
		})
	}
}

func TestDoBadResponseAndLog_JsonMarshalError(t *testing.T) {
	w := httptest.NewRecorder()

	originalMarshal := jsonMarshal

	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("marshal error")
	}

	defer func() {
		jsonMarshal = originalMarshal
	}()

	DoBadResponseAndLog(w, http.StatusBadRequest, "test")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "internal error")
}

func TestDoJSONResponse_Success(t *testing.T) {
	tests := []struct {
		name              string
		responseData      any
		successStatusCode int
		expectedBody      string
	}{
		{
			name:              "SimpleString",
			responseData:      map[string]string{"message": "success"},
			successStatusCode: http.StatusOK,
			expectedBody:      `{"message":"success"}`,
		},
		{
			name:              "ComplexStruct",
			responseData:      BadResponse{Status: 200, Text: "OK"},
			successStatusCode: http.StatusCreated,
			expectedBody:      `{"status":200,"text":"OK"}`,
		},
		{
			name:              "EmptyStruct",
			responseData:      struct{}{},
			successStatusCode: http.StatusNoContent,
			expectedBody:      `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			DoJSONResponse(w, tt.responseData, tt.successStatusCode)

			assert.Equal(t, tt.successStatusCode, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			expectedLength := strconv.Itoa(len(tt.expectedBody))
			assert.Equal(t, expectedLength, w.Header().Get("Content-Length"))

			assert.JSONEq(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestDoJSONResponse_MarshalError(t *testing.T) {
	w := httptest.NewRecorder()

	invalidData := make(chan int)

	DoJSONResponse(w, invalidData, http.StatusOK)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response BadResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Status)
	assert.Equal(t, "internal error", response.Text)
}

func TestDoJSONResponse_WriteError(t *testing.T) {
	mockWriter := &mockResponseWriter{
		headers: make(http.Header),
	}

	DoJSONResponse(mockWriter, map[string]string{"test": "data"}, http.StatusOK)

	assert.Equal(t, "application/json", mockWriter.headers.Get("Content-Type"))
}

func TestResponseErrorAndLog(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		funcName       string
		expectedStatus int
		expectedText   string
	}{
		{
			name:           "NotFoundError",
			err:            errs.ErrNotFound,
			funcName:       "GetOrderByID",
			expectedStatus: http.StatusNotFound,
			expectedText:   "order not found",
		},
		{
			name:           "ValidationError",
			err:            errs.ErrValidation,
			funcName:       "CreateOrder",
			expectedStatus: http.StatusBadRequest,
			expectedText:   "invalid request data",
		},
		{
			name:           "GenericError",
			err:            errors.New("some random error"),
			funcName:       "ProcessOrder",
			expectedStatus: http.StatusInternalServerError,
			expectedText:   "internal server error",
		},
		{
			name:           "WrappedNotFoundError",
			err:            fmt.Errorf("wrapped: %w", errs.ErrNotFound),
			funcName:       "GetOrder",
			expectedStatus: http.StatusNotFound,
			expectedText:   "order not found",
		},
		{
			name:           "WrappedValidationError",
			err:            fmt.Errorf("wrapped: %w", errs.ErrValidation),
			funcName:       "ValidateOrder",
			expectedStatus: http.StatusBadRequest,
			expectedText:   "invalid request data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			ResponseErrorAndLog(w, tt.err, tt.funcName)

			var response BadResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, response.Status)
			assert.Equal(t, tt.expectedText, response.Text)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestBadResponse_JSONEncoding(t *testing.T) {
	br := BadResponse{
		Status: 404,
		Text:   "Not Found",
	}

	jsonData, err := json.Marshal(br)
	assert.NoError(t, err)

	var decoded BadResponse
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, br.Status, decoded.Status)
	assert.Equal(t, br.Text, decoded.Text)
}

type mockResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = data
	return 0, errors.New("write error")
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}
