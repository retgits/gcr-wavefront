package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	gcrwavefront "github.com/retgits/gcr-wavefront"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func main() {
	log.Print("Hello world sample started.")

	cfg := gcrwavefront.WavefrontConfig{
		Server:        "https://try.wavefront.com",
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

	router := httprouter.New()
	router.GET("/", cfg.WrapHTTPHandle(Index))
	router.GET("/hello/:name", cfg.WrapHTTPHandle(Hello))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
