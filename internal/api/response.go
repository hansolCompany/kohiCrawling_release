package api

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, body response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeTaskSuccess(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, response{Status: "success"})
}

func writeTaskError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusOK, response{
		Status:  "error",
		Message: message,
	})
}

func writeStarted(w http.ResponseWriter) {
	writeJSON(w, http.StatusAccepted, response{Status: "started"})
}

func writeRunning(w http.ResponseWriter) {
	writeJSON(w, http.StatusConflict, response{
		Status:  "running",
		Message: "해당 크롤링이 이미 실행 중입니다",
	})
}

func writeError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, response{
		Status:  "error",
		Message: message,
	})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, response{
		Status:  "error",
		Message: "POST만 허용됩니다",
	})
}

func writeJSONData(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
