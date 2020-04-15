package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	gcrwavefront "github.com/retgits/gcr-wavefront"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("Hello world received a request.")
	target := os.Getenv("TARGET")
	if target == "" {
		target = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)
}

func main() {
	log.Print("Hello world sample started.")

	cfg := gcrwavefront.WavefrontConfig{
		Server:        "https://try.wavefront.com",
		Token:         "0731ca22-2e9f-4ea5-ba62-c210de96c9c9",
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

	http.HandleFunc("/", cfg.WrapHandlerFunc(handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
