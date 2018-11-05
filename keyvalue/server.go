package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	openlog "github.com/opentracing/opentracing-go/log"
)

var db *pool.Pool

var (
	RedisUrl = "localhost:6379"
)

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

	startRedis()

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/keyvalue/{book}/{canto}",  searchHandler)
	r.HandleFunc("/api/v1/keyvalue/{book}/{canto}",  postHandler).Methods("POST")

	panic(http.ListenAndServe(":"+port, r))
}

func startRedis() {
	var err error
	// Establish a pool of 10 connections to the Redis server listening on
	// port 6379 of the variable that has been used
	if os.Getenv("REDIS_URL") != "" {
		RedisUrl = os.Getenv("REDIS_URL")
	}
	db, err = pool.New("tcp", RedisUrl, 15)
	if err != nil {
	}
}

func searchHandler(w http.ResponseWriter, req *http.Request) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("postHandler")
	span.SetTag("Method", "postHandler")
	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
	)

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	key := "inferno:cantoi"
	testSting := "this is a test string with words words words words word s"
	words := strings.Fields(testSting)
	delWordCount(key, ctx)
	for _, word := range words{
		incrWordCount(key, word, ctx)
	}

	c := getWordCount(key, ctx)
	respondWithJson(w, 200, c, ctx)
}

func postHandler(w http.ResponseWriter, req *http.Request) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("postHandler")
	span.SetTag("Method", "postHandler")
	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
	)

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	key := "inferno:cantoi"
	testSting := "this is a test string with words words words words word s"
	words := strings.Fields(testSting)
	delWordCount(key, ctx)
	for _, word := range words{
		incrWordCount(key, word, ctx)
	}
}


func delWordCount(key string, ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "delWordCount")
	span.SetTag("Method", "delWordCount")

	defer span.Finish()

	res, err := db.Cmd("EXISTS", key).Int()
	if err != nil {
		println(err)
	}
	exists := res != 0

	if exists {
		db.Cmd("DEL", key)
	}
}
func incrWordCount(key string, word string, ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "incrWordCount")
	span.SetTag("Method", "incrWordCount")

	defer span.Finish()

	err := db.Cmd("HINCRBY", key, word, 1).Err
	if err != nil {
		println(err)
	}
	span.LogFields(
		openlog.String("http_status_code", "200"),
		openlog.String("body", "Increased " + word + " by 1"),
	)
	return
}

func getWordCount(key string, ctx context.Context) (map[string] string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "incrWordCount")
	span.SetTag("Method", "incrWordCount")

	defer span.Finish()

	result, err := db.Cmd("HGETALL", key).Map()
	if err != nil {
		println(err)
	}

	span.LogFields(
		openlog.String("http_status_code", "200"),
		openlog.String("body", "Increased by 1"),
	)
	return result
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

//generic method to respondwith json and log to jaeger
func respondWithJson(w http.ResponseWriter, code int, payload interface{}, ctx context.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Response")
	span.SetTag("Method", "ShowHighScores")

	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	span.LogFields(
		openlog.String("http_status_code", strconv.Itoa(code)),
		openlog.String("body", string(response)),
	)
	defer span.Finish()
}

