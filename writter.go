package session

import (
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	cookie        *http.Cookie
	statusWritten bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusWritten {
		return
	}

	if rw.cookie != nil {
		http.SetCookie(rw, rw.cookie)
	}

	rw.statusWritten = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) AddCookie(c *http.Cookie) {
	rw.cookie = c
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.statusWritten {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
