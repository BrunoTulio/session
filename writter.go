package session

import (
	"net/http"
)

type responseWriter struct {
	http.ResponseWriter
	cookies       []*http.Cookie
	statusWritten bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusWritten {
		return
	}

	for _, c := range rw.cookies {
		http.SetCookie(rw, c)
	}

	rw.statusWritten = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) AddCookie(c *http.Cookie) {
	rw.cookies = append(rw.cookies, c)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.statusWritten {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
