package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"net/http"
	"os"
)

var db *pool.Pool

func init() {
	var err error
	// Establish a pool of 10 connections to the Redis server listening on
	// port 6379 of the variable that has been used
	redisUrl := os.Getenv("REDIS_URL") + ":6379"
	fmt.Println(redisUrl)
	db, err = pool.New("tcp", redisUrl, 10)
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

	helloTo := "Redis"

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	helloStr := formatString(ctx, helloTo)
	printHello(ctx, helloStr)
	span.Finish()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3230"
	}

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

func formatString(ctx context.Context, helloTo string) string {
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()
	helloStr := fmt.Sprintf("Hello, %s!", helloTo)
	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)

	return helloStr
}

func printHello(ctx context.Context, helloStr string){
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	println(helloStr)
	span.LogKV("event", "println")
}



