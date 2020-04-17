package gcrwavefront

import (
	"net/http"
	"time"
)

// WrapHandlerFunc is the wrapper needed for use with the net/http package
func (wc *WavefrontConfig) WrapHandlerFunc(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()
		// Initialize the status to 200 in case WriteHeader is not called
		rec := &WFStatusRecorder{w, 200, 0}
		defer emitMetrics(wc, rec, r, start)
		f(rec, r)
	}
}
