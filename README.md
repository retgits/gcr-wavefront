# gcr-wavefront

[![Go Report Card](https://goreportcard.com/badge/github.com/retgits/gin-wavefront?style=flat-square)](https://goreportcard.com/report/github.com/retgits/gcr-wavefront)
[![Godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/retgits/gcr-wavefront)
![GitHub](https://img.shields.io/github/license/retgits/gcr-wavefront?style=flat-square)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/retgits/gcr-wavefront?sort=semver&style=flat-square)

> gcr-wavefront is an HTTP middleware to emit Google Cloud Run metrics to [Wavefront](https://www.wavefront.com/).

## Prerequisites

To use this HTTP middleware, you'll need to have

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [A Wavefront account](https://www.wavefront.com/sign-up/)
* [A Wavefront API key](https://docs.wavefront.com/wavefront_api.html)

## Installation

Using `go get`

```bash
go get github.com/retgits/gcr-wavefront
```

## Usage

To start, you'll need to initialize the Wavefront emitter:

```go
// Set configuration parameters
wfconfig := &gcrwavefront.WavefrontConfig{
    Server:        "https://<INSTANCE>.wavefront.com",
    Token:         "my-api-key",
    BatchSize:     10000,
    MaxBufferSize: 50000,
    FlushInterval: 1,
    Source:        "my-app",
    MetricPrefix:  "my.awesome.app",
    PointTags:     make(map[string]string),
}

// Make sure the sender is configured and initialized
err := cfg.ConfigureSender()
if err != nil {
    panic(err)
}

// Wrap your handler
http.HandleFunc("/", cfg.WrapHandlerFunc(handler))
```

A complete sample app can be found in the [examples](./examples) folder

## Contributing

[Pull requests](https://github.com/retgits/gcr-wavefront/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/gcr-wavefront/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
