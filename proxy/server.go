package main

import (
	"context"
	"fmt"
	microclient "github.com/joerivrij/microbases/shared/client"
	"github.com/joerivrij/microbases/shared/response"
	"github.com/joerivrij/microbases/shared/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"html/template"
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

	fs := http.FileServer(http.Dir("proxy/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", HomePage)
	http.HandleFunc("/queryWordCount", QueryWordCount)
	http.ListenAndServe(":3201", nil)
}

func render(w http.ResponseWriter, tmpl string) {
	tmpl = fmt.Sprintf("proxy/templates/%s", tmpl) // prefix the name passed in with templates/
	t, err := template.ParseFiles(tmpl)      //parse the template file held in the templates folder

	if err != nil {
		fmt.Print("template parsing error: ", err)
	}

	err = t.Execute(w, nil) //execute the template and pass in the variables to fill the gaps

	if err != nil {
		fmt.Print("template executing error: ", err)
	}
}

func HomePage(w http.ResponseWriter, r *http.Request){
	render(w, "home.html")
}

func QueryWordCount(w http.ResponseWriter, req *http.Request)  {
	span := opentracing.GlobalTracer().StartSpan("queryWordCount")
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	token := GetToken(ctx, req)

	span.LogFields(
		log.String("event", "retrieved a token"),
		log.String("value", token),
	)

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

	resp, err := microclient.BackendCall(req)
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


func GetToken(ctx context.Context, req *http.Request) string {
	span, _ := opentracing.StartSpanFromContext(ctx, "getToken")
	tokenUrl := "http://localhost:3240/token?grant_type=client_credentials&client_id=000000&client_secret=999999&scope=read"
	req, err := http.NewRequest("GET", tokenUrl, nil)
	if err != nil {
		panic(err.Error())
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, tokenUrl)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)

	respToken, err := microclient.BackendCall(req)
	if err != nil {
		panic(err.Error())
	}

	tokenString := string(respToken)
	span.LogFields(
		log.String("event", "calling oauth server"),
		log.String("value", tokenString),
	)
	return tokenString
}

