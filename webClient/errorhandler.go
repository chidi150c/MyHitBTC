package webClient

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

// Error writes an API error message to the response and logger.
func Error(w http.ResponseWriter, err error, code int, logger *log.Logger) {
	// Log error.
	logger.Printf("http error: %s (code=%d) Error: internal error", err, code)
	fmt.Println()
	// Hide error from client if it is internal.
	if code == http.StatusInternalServerError {
		err = errors.New("internal error")
	}

	// Write generic error response.
	http.Error(w, fmt.Sprintf("Error: %+v", err.Error()), code)
}

// NotFound writes an API error message to the response.
func NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{}` + "\n"))
}
