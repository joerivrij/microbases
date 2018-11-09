package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joerivrij/microbases/shared/models"
	"github.com/joerivrij/microbases/shared/response"
	"github.com/joerivrij/microbases/shared/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	openlog "github.com/opentracing/opentracing-go/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	tracer, closer := tracing.Init("DocumentBackendApi")
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
	tracing.PrintServerInfo(ctx, logValue)
	span.Finish()

	r := mux.NewRouter()
	r.HandleFunc("/api/{book}/canti", allCantiHandler).Methods("GET")
	r.HandleFunc("/api/{book}/{canto}", specificCantoHandler).Methods("GET")
	r.HandleFunc("/api/{book}/{canto}/{verse}", specificCantoWithVerseHandler).Methods("GET")

	panic(http.ListenAndServe(":"+port, r))
}

func specificCantoHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	book := strings.Title(vars["book"])
	canto := vars["canto"]

	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := opentracing.GlobalTracer().StartSpan("postHandler", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
	)
	arabic, err := strconv.Atoi(canto)
	if err != nil {
		fmt.Println(err)
	}
	query := bson.M{"book": book, "arabic": arabic}

	result := findAllWithQuery(ctx, query)

	response.RespondWithJson(w, http.StatusOK, result, ctx)
}

func specificCantoWithVerseHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	book := strings.Title(vars["book"])
	canto := vars["canto"]
	verse := vars["verse"]

	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := opentracing.GlobalTracer().StartSpan("postHandler", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
	)

	arabic, err := strconv.Atoi(canto)
	if err != nil {
		fmt.Println(err)
	}

	verseBson, err := strconv.Atoi(verse)
	if err != nil {
		fmt.Println(err)
	}
	query := bson.M{"book": book, "arabic": arabic, "verse": verseBson}
	result := findOneWithQuery(ctx, query)

	response.RespondWithJson(w, http.StatusOK, result, ctx)
}

func allCantiHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	book := strings.Title(vars["book"])

	spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	span := opentracing.GlobalTracer().StartSpan("postHandler", ext.RPCServerOption(spanCtx))
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	span.LogFields(
		openlog.String("method", req.Method),
		openlog.String("path", req.URL.Path),
		openlog.String("host", req.Host),
	)

	canti, err := findAll(ctx, book)
	if err != nil {
		//respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.RespondWithJson(w, http.StatusOK, canti, ctx)

	}

func findAll(ctx context.Context, book string) ([]models.Canto, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "findAll")
	defer span.Finish()
	var canti []models.Canto

	err := db.C(COLLECTION).Find(bson.M{"book": book}).All(&canti)
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

func findAllWithQuery(ctx context.Context, query bson.M) ([]models.Canto) {
	span, _ := opentracing.StartSpanFromContext(ctx, "findWithQuery")
	defer span.Finish()
	var canti []models.Canto

	err := db.C(COLLECTION).Find(query).All(&canti)
	if err != nil {
		span.LogFields(
			openlog.String("mongoresult", "error getting canti"),
		)
	}

	response, _ := json.Marshal(canti)
	span.LogFields(
		openlog.String("mongoresult", string(response)),
	)

	return canti
}

func findOneWithQuery(ctx context.Context, query bson.M) (models.Canto) {
	span, _ := opentracing.StartSpanFromContext(ctx, "findWithQuery")
	defer span.Finish()
	var canto models.Canto

	err := db.C(COLLECTION).Find(query).One(&canto)
	if err != nil {
		span.LogFields(
			openlog.String("mongoresult", "error getting canti"),
		)
	}

	response, _ := json.Marshal(canto)
	span.LogFields(
		openlog.String("mongoresult", string(response)),
	)

	return canto
}