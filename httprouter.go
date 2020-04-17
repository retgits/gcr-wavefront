package gcrwavefront

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// WrapHTTPHandle is the wrapper needed for use with the httprouter project (https://github.com/julienschmidt/httprouter)
func (wc *WavefrontConfig) WrapHTTPHandle(f httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// Start timer
		start := time.Now()
		// Initialize the status to 200 in case WriteHeader is not called
		rec := &WFStatusRecorder{w, 200, 0}
		defer emitMetrics(wc, rec, r, start)
		f(rec, r, p)
	}
}
