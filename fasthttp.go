package gcrwavefront

import (
	"time"

	"github.com/valyala/fasthttp"
)

// WrapFastHTTPRequest is the wrapper needed for use with the fasthttp project (https://github.com/valyala/fasthttp)
func (wc *WavefrontConfig) WrapFastHTTPRequest(f fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Start timer
		start := time.Now()
		defer wc.buildFastHTTPMetrics(ctx, start)
		f(ctx)
	}
}

func (wc *WavefrontConfig) buildFastHTTPMetrics(ctx *fasthttp.RequestCtx, start time.Time) {
	// Add tags
	wc.PointTags["path"] = string(ctx.URI().Path())
	wc.PointTags["method"] = string(ctx.Method())
	wc.PointTags["userAgent"] = string(ctx.UserAgent())

	// Stop timer and emit metrics
	end := time.Now()
	wc.emitMetrics(metrics{
		EndTime:        end,
		Latency:        end.Sub(start),
		HTTPStatusCode: ctx.Response.StatusCode(),
		BytesOut:       len(ctx.Response.Body()),
		BytesIn:        len(ctx.Request.Body()),
	})
}
