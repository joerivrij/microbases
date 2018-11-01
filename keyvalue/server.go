package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"net/http"
	"os"
)

var db *pool.Pool

var (
	RedisUrl = "http://localhost:6379"
)

func init() {
	var err error
	// Establish a pool of 10 connections to the Redis server listening on
	// port 6379 of the variable that has been used
	if os.Getenv("REDIS_URL") != "" {
		RedisUrl = os.Getenv("REDIS_URL")
	}
	db, err = pool.New("tcp", RedisUrl, 10)
	if err != nil {
	}
}

func searchHandler(w http.ResponseWriter, req *http.Request) {
}

func postHandler(w http.ResponseWriter, req *http.Request) {
}

func main() {
	tracer, closer := initJaeger("KeyValueBackendApi")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3230"
	}

	span := tracer.StartSpan("StartingKeyValueServer")
	span.SetTag("event", "Starting MUX")
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	logValue := fmt.Sprintf("Starting server on port %s with redis %s", port, RedisUrl)
	printServerInfo(ctx, logValue)
	span.Finish()

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/redis/{user}/{playList}",  searchHandler)
	r.HandleFunc("/api/v1/redis/{user}/{playList}",  postHandler).Methods("POST")

	panic(http.ListenAndServe(":"+port, r))
}

// initJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}



func printServerInfo(ctx context.Context, serverInfo string){
	span, _ := opentracing.StartSpanFromContext(ctx, "ServerInfo")
	defer span.Finish()

	span.LogKV("event", serverInfo)
}



