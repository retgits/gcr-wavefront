package gcrwavefront

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

const (
	// ErrCreateSender in case any errors occur while creating the Wavefront Direct Sender
	ErrCreateSender = "error creating wavefront sender: %s"
	// DebugServerName has the server name to set when you want to print things to the log instead of sending data to Wavefront
	DebugServerName = "debug"
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

type metrics struct {
	EndTime        time.Time
	Latency        time.Duration
	HTTPStatusCode int
	BytesOut       int
	BytesIn        int
}

// WavefrontConfig configures the direct ingestion sender to Wavefront.
type WavefrontConfig struct {
	// Wavefront URL of the form https://<INSTANCE>.wavefront.com.
	// Setting the server to debug will print the metrics to a log instead of sending them to Wavefront
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

func emitMetrics(wc *WavefrontConfig, rec *WFStatusRecorder, r *http.Request, start time.Time) {
	// Add tags
	wc.PointTags["path"] = r.URL.Path
	wc.PointTags["method"] = r.Method
	wc.PointTags["userAgent"] = r.UserAgent()

	// Stop timer and emit metrics
	end := time.Now()
	wc.emitMetrics(metrics{
		EndTime:        end,
		Latency:        end.Sub(start),
		HTTPStatusCode: rec.status,
		BytesOut:       rec.size,
		BytesIn:        int(r.ContentLength),
	})
}

func (wc *WavefrontConfig) emitMetrics(m metrics) {
	// Print to log
	if wc.Server == DebugServerName {
		log.Printf("Tags: %+v\nDuration: %+v\nStatus: %d\nBytesOut: %d\nBytesIn: %d", wc.PointTags, m.Latency.Microseconds(), m.HTTPStatusCode, m.BytesOut, m.BytesIn)
		return
	}

	// Send metrics
	sender.SendMetric(strings.Join([]string{wc.MetricPrefix, ".latency"}, ""), float64(m.Latency.Microseconds()), m.EndTime.Unix(), wc.Source, wc.PointTags)
	sender.SendMetric(strings.Join([]string{wc.MetricPrefix, ".bytes.in"}, ""), float64(m.BytesIn), m.EndTime.Unix(), wc.Source, wc.PointTags)
	sender.SendMetric(strings.Join([]string{wc.MetricPrefix, ".bytes.out"}, ""), float64(m.BytesOut), m.EndTime.Unix(), wc.Source, wc.PointTags)
	switch {
	case m.HTTPStatusCode > 199 && m.HTTPStatusCode < 300:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.success"}, ""), 1, wc.Source, wc.PointTags)
	case m.HTTPStatusCode > 299 && m.HTTPStatusCode < 400:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.redirection"}, ""), 1, wc.Source, wc.PointTags)
	case m.HTTPStatusCode > 399 && m.HTTPStatusCode < 500:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.error.client"}, ""), 1, wc.Source, wc.PointTags)
	case m.HTTPStatusCode > 499 && m.HTTPStatusCode < 600:
		sender.SendDeltaCounter(strings.Join([]string{wc.MetricPrefix, ".status.error.server"}, ""), 1, wc.Source, wc.PointTags)
	}
}
