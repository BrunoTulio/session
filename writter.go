package session

import (
	"net/http"
	"sync"
)

type responseWriter struct {
	http.ResponseWriter
	statusWritten bool
	once          sync.Once
	before        func()
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusWritten {
		return
	}

	if rw.before != nil {
		rw.once.Do(rw.before)
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
