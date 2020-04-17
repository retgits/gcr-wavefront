package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	gcrwavefront "github.com/retgits/gcr-wavefront"
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "Welcome!\n")
}

func Hello(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "hello, %s!\n", ctx.UserValue("name"))
}

func main() {
	log.Print("Hello world sample started.")

	cfg := gcrwavefront.WavefrontConfig{
		Server:        gcrwavefront.DebugServerName,
		Token:         "my-api-key",
		BatchSize:     10000,
		MaxBufferSize: 50000,
		FlushInterval: 1,
		Source:        "my-app",
		MetricPrefix:  "my.awesome.app",
		PointTags:     make(map[string]string),
	}
	err := cfg.ConfigureSender()
	if err != nil {
		panic(err)
	}

	router := router.New()
	router.GET("/", cfg.WrapFastHTTPRequest(Index))
	router.GET("/hello/:name", cfg.WrapFastHTTPRequest(Hello))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(fasthttp.ListenAndServe(fmt.Sprintf(":%s", port), router.Handler))
}
