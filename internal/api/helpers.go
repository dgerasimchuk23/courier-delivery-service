package api

import "net/http"

func writeError(w http.ResponseWriter, message string, statusCode int) {
	http.Error(w, message, statusCode)
}
