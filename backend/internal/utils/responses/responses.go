package responses

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/supchaser/wb_l0/internal/utils/logger"
	"go.uber.org/zap"
)

type BadResponse struct {
	Status int    `json:"status"`
	Text   string `json:"text"`
}

func DoBadResponseAndLog(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := BadResponse{
		Status: statusCode,
		Text:   message,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonResponse)
	if err != nil {
		logger.Error("failed to write response",
			zap.String("function", "DoBadResponseAndLog"),
			zap.Error(err),
		)
		return
	}

	logger.Warn("Bad response",
		zap.Int("status", statusCode),
		zap.String("message", message),
	)
}

func DoJSONResponse(w http.ResponseWriter, responseData interface{}, successStatusCode int) {
	body, err := json.Marshal(responseData)
	if err != nil {
		DoBadResponseAndLog(w, http.StatusInternalServerError, "internal error")
		logger.Error("failed to marshal response",
			zap.String("function", "DoJSONResponse"),
			zap.Error(err),
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(successStatusCode)

	if _, err := w.Write(body); err != nil {
		logger.Error("failed to write response",
			zap.String("function", "DoJSONResponse"),
			zap.Error(err),
		)
	}
}

func ResponseErrorAndLog(w http.ResponseWriter, err error, funcName string) {
	switch {

	default:
		DoBadResponseAndLog(w, http.StatusInternalServerError, "internal error")
		logger.Error(funcName,
			zap.String("error", err.Error()),
		)
	}
}
