package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joerivrij/microbases/document/models"
	"github.com/opentracing/opentracing-go"
	openlog "github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"net/http"
	"os"
	"strconv"
)


var db *mgo.Database

const (
	COLLECTION = "canti"
	SERVER = "localhost:27017"
	DATABASE = "divinacommedia"
)

func Connect() {
	session, err := mgo.Dial(SERVER)
	if err != nil {
		println(err)
	}
	db = session.DB(DATABASE)
}

func main() {
	Connect()
	tracer, closer := initJaeger("DocumentBackendApi")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	span := tracer.StartSpan("StartingServer")
	span.SetTag("event", "Starting MUX")
	defer span.Finish()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3210"
	}

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	logValue := fmt.Sprintf("Starting server on port %s with mongodb %s", port, db)
	printServerInfo(ctx, logValue)
	span.Finish()

	r := mux.NewRouter()
	r.HandleFunc("/api/canti", allCantiHandler).Methods("GET")

	panic(http.ListenAndServe(":"+port, r))
}

func allCantiHandler(w http.ResponseWriter, req *http.Request) {
	tracer := opentracing.GlobalTracer()
	span := tracer.StartSpan("allCantiHandler")
	span.SetTag("Method", "allCantiHandler")
	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
	)
	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)
	defer span.Finish()

	canti, err := findAll(ctx)
	if err != nil {
		//respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, canti, ctx)

	}

func findAll(ctx context.Context) ([]models.Canto, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "findAll")
	defer span.Finish()
	var canti []models.Canto

	err := db.C(COLLECTION).Find(bson.M{}).All(&canti)
	if err != nil {
		span.LogFields(
			openlog.String("mongoresult", "error getting canti"),
		)
	}

	response, _ := json.Marshal(canti)
	span.LogFields(
		openlog.String("mongoresult", string(response)),
	)

	return canti, err
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


// startup log with server info
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