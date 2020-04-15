package acmeserverless

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

const (
	// ErrCreateSender in case any errors occur while creating the Wavefront Direct Sender
	ErrCreateSender = "error creating wavefront sender: %s"
)

var sender wavefront.Sender

type WFStatusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (rec *WFStatusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *WFStatusRecorder) Write(payload []byte) (int, error) {
	size, err := rec.ResponseWriter.Write(payload)
	rec.size = size
	return size, err
}

// WavefrontConfig configures the direct ingestion sender to Wavefront.
type WavefrontConfig struct {
	// Wavefront URL of the form https://<INSTANCE>.wavefront.com.
	Server string
	// Wavefront API token with direct data ingestion permission.
	Token string
	// Max batch of data sent per flush interval.
	BatchSize int
	// Max batch of data sent per flush interval.
	MaxBufferSize int
	// Interval (in seconds) at which to flush data to Wavefront.
	FlushInterval int
	// Map of Key-Value pairs (strings) associated with each data point sent to Wavefront.
	PointTags map[string]string
	// Name of the app that emits metrics.
	Source string
	// Prefix added to all metrics
	MetricPrefix string
}

// ConfigureSender ...
func (w *WavefrontConfig) ConfigureSender() error {
	if w.PointTags == nil {
		w.PointTags = make(map[string]string)
	}

	dc := &wavefront.DirectConfiguration{
		Server:               w.Server,
		Token:                w.Token,
		BatchSize:            w.BatchSize,
		MaxBufferSize:        w.MaxBufferSize,
		FlushIntervalSeconds: w.FlushInterval,
	}

	var err error

	sender, err = wavefront.NewDirectSender(dc)
	if err != nil {
		return fmt.Errorf(ErrCreateSender, err.Error())
	}

	return nil
}

// WrapHandlerFunc ...
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

func emitMetrics(wc *WavefrontConfig, rec *WFStatusRecorder, r *http.Request, start time.Time) {
	// Stop timer
	end := time.Now()
	latency := end.Sub(start)
	statusCode := rec.status
	bytesOut := rec.size
	bytesIn := r.ContentLength

	// Add tags
	wc.PointTags["path"] = r.URL.Path
	wc.PointTags["clientIP"] = getClientIP(r)
	wc.PointTags["method"] = r.Method
	wc.PointTags["userAgent"] = r.UserAgent()

	// Send metrics
	// <metricName> <metricValue> [<timestamp>] source=<source> [pointTags]
	sender.SendMetric(strings.Join([]string{wc.MetricPrefix, ".latency"}, ""), float64(latency.Milliseconds()), end.Unix(), wc.Source, wc.PointTags)
	sender.SendMetric(strings.Join([]string{wc.MetricPrefix, ".bytes.in"}, ""), float64(bytesIn), end.Unix(), wc.Source, wc.PointTags)
	sender.SendMetric(strings.Join([]string{wc.MetricPrefix, ".bytes.out"}, ""), float64(bytesOut), end.Unix(), wc.Source, wc.PointTags)
	switch {
	case statusCode > 199 && statusCode < 300:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.success"}, ""), 1, wc.Source, wc.PointTags)
	case statusCode > 299 && statusCode < 400:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.redirection"}, ""), 1, wc.Source, wc.PointTags)
	case statusCode > 399 && statusCode < 500:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.error.client"}, ""), 1, wc.Source, wc.PointTags)
	case statusCode > 499 && statusCode < 600:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.error.server"}, ""), 1, wc.Source, wc.PointTags)
	}

	// DEBUG
	//fmt.Printf("Tags: %+v\nDuration: %+v\nStatus: %d\nBytesOut: %d\nBytesIn: %d", wc.PointTags, latency, statusCode, bytesOut, bytesIn)
}

// getClientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
// Use X-Forwarded-For before X-Real-Ip as nginx uses X-Real-Ip with the proxy's IP.
// source: https://github.com/gin-gonic/gin/blob/master/context.go
func getClientIP(r *http.Request) string {
	clientIP := r.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if clientIP == "" {
		clientIP = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
		if clientIP == "" {
			if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
				clientIP = ip
			}
		}
	}

	return clientIP
}
