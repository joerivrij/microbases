package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joerivrij/microbases/shared/tracing"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	openlog "github.com/opentracing/opentracing-go/log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var db *pool.Pool

var (
	RedisUrl = "localhost:6379"
)

type PostBody struct {
	Words  string `json:"words"`
}

func main() {
	jaegerUrl := os.Getenv("JAEGER_AGENT_HOST")
	jaegerPort :=  os.Getenv("JAEGER_AGENT_PORT")
	jaegerConfig := jaegerUrl + ":" + jaegerPort
	println(jaegerConfig)

	tracer, closer := tracing.Init("KeyValueBackendApi", jaegerConfig)
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
	r.HandleFunc("/api/v1/keyvalue/{book}/{canto}/{verse}", searchHandler).Methods("GET")
	r.HandleFunc("/api/v1/keyvalue/{book}/{canto}/{verse}", postHandler).Methods("POST")

	panic(http.ListenAndServe(":"+port, r))
}

func startRedis() {
	var err error
	// Establish a pool of 15 connections to the Redis server listening on
	// port 6379 of the variable that has been used
	if os.Getenv("REDIS_URL") != "" {
		RedisUrl = os.Getenv("REDIS_URL")
	}
	db, err = pool.New("tcp", RedisUrl, 15)
	if err != nil {
	}
}

func searchHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	book := vars["book"]
	canto := vars["canto"]
	verse := vars["verse"]

	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := opentracing.GlobalTracer().StartSpan("searchHandler", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
	)

	key := book + ":" + canto + ":" + verse

	exists := keyExists(key, ctx)
	if !exists{
		respondWithJson(w, 503, "not done yet", ctx)
	}

	c := getWordCount(key, ctx)
	respondWithJson(w, 200, c, ctx)
}

func postHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	book := vars["book"]
	canto := vars["canto"]
	verse := vars["verse"]

	var m PostBody

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJson(w, http.StatusBadRequest, req.Body, context.Background())
		return
	}

	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := opentracing.GlobalTracer().StartSpan("postHandler", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	jsonBody, _ := json.Marshal(m)
	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
		openlog.String("body", string(jsonBody)),
	)

	key := book + ":" + canto + ":" + verse
	words := strings.Fields(m.Words)
	delWordCount(key, ctx)
	for _, word := range words{
		incrWordCount(key, word, ctx)
	}

	w.WriteHeader(201)
	respondWithJson(w, 201, "Created", ctx)
}

func keyExists(key string, ctx context.Context) (bool) {
	span, _ := opentracing.StartSpanFromContext(ctx, "keyExists")
	span.SetTag("Method", "keyExists")

	defer span.Finish()

	res, err := db.Cmd("EXISTS", key).Int()
	if err != nil {
		println(err)
	}
	exists := res != 0

	return exists
}

func getWordCountFromMongo(key string, ctx context.Context) {

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
	span, _ := opentracing.StartSpanFromContext(ctx, "GetWordCount")
	span.SetTag("Method", "GetWordCount")

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

