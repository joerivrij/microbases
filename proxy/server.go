package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joerivrij/microbases/shared/response"
	"github.com/joerivrij/microbases/shared/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	tracer, closer := tracing.Init("ProxyApi")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3201"
	}

	span := tracer.StartSpan("StartingProxyApi")
	span.SetTag("event", "Starting MUX")
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	logValue := fmt.Sprintf("Starting server on port %s", port)
	tracing.PrintServerInfo(ctx, logValue)
	span.Finish()

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/proxy", queryWordCount).Methods("GET")
	r.HandleFunc("/api/v1/proxy/{book}/{canto}/{verse}", postHandler).Methods("POST")
	r.HandleFunc("/api/v1/proxy/{book}/{canto}/{verse}", getTextToQuery).Methods("GET")

	panic(http.ListenAndServe(":"+port, r))
}

func getTextToQuery(w http.ResponseWriter, req *http.Request) {

}

func queryWordCount(w http.ResponseWriter, req *http.Request)  {
	span := opentracing.GlobalTracer().StartSpan("queryWordCount")
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	url := "http://localhost:3230/api/v1/keyvalue/inferno/cantoi/1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	resp, err := Do(req)
	if err != nil {
		panic(err.Error())
	}

	helloStr := string(resp)
	span.LogFields(
		log.String("event", "calling backend"),
		log.String("value", helloStr),
	)

	response.RespondWithJson(w, 200, helloStr, ctx)
}


func Do(req *http.Request) ([]byte, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode: %d, Body: %s", resp.StatusCode, body)
	}

	return body, nil
}
