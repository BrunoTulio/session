package session

import (
	"net/http"
)

type handler interface {
	commit(w http.ResponseWriter, r *http.Request) error
	onError(w http.ResponseWriter, r *http.Request, err error)
}

type responseWriter struct {
	http.ResponseWriter
	request       *http.Request
	statusWritten bool
	handler       handler
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusWritten {
		return
	}

	if err := rw.handler.commit(rw.ResponseWriter, rw.request); err != nil {
		rw.statusWritten = true
		rw.handler.onError(rw.ResponseWriter, rw.request, err)
		return
	}

	rw.statusWritten = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.statusWritten {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}
