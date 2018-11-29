package main

import (
	"context"
	"fmt"
	microclient "github.com/joerivrij/microbases/shared/client"
	"github.com/joerivrij/microbases/shared/response"
	"github.com/joerivrij/microbases/shared/tracing"
	"github.com/joho/godotenv"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	openlog "github.com/opentracing/opentracing-go/log"
	"html/template"
	"log"
	"net/http"
	"os"
)


var (
	DocumentUrl = "localhost:3210"
)

var (
	KeyvalueUrl = "localhost:3230"
)

var (
	OauthUrl = "localhost:3240"
)

var (
	GraphUrl =  "localhost:3220"
)

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
		openlog.String("event", "retrieved a token"),
		openlog.String("value", token),
	)

	url := fmt.Sprintf("http://%s/api/v1/keyvalue/inferno/cantoi/1", KeyvalueUrl)
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
		openlog.String("event", "calling backend"),
		openlog.String("value", helloStr),
	)

	response.RespondWithJson(w, 200, helloStr, ctx)
}


func GetToken(ctx context.Context, req *http.Request) string {
	span, _ := opentracing.StartSpanFromContext(ctx, "getToken")
	tokenUrl := fmt.Sprintf("http://%s/token?grant_type=client_credentials&client_id=000000&client_secret=999999&scope=read", OauthUrl)
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
		openlog.String("event", "calling oauth server"),
		openlog.String("value", tokenString),
	)
	return tokenString
}

func init() {
	systemEnv := os.Getenv("GOENV")
	err := godotenv.Load(".env." + systemEnv)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	println(os.Getenv("DOCUMENT_URL"))

	DocumentUrl = os.Getenv("DOCUMENT_URL")
	KeyvalueUrl = os.Getenv("KEYVALUE_URL")
	OauthUrl = os.Getenv("OAUTH_URL")
	GraphUrl = os.Getenv("GRAPH_URL")

	println(DocumentUrl)
	println(KeyvalueUrl)
	println(OauthUrl)
	println(GraphUrl)
}

func main() {
	jaegerUrl := os.Getenv("JAEGER_AGENT_HOST")
	jaegerPort :=  os.Getenv("JAEGER_AGENT_PORT")
	jaegerConfig := jaegerUrl + ":" + jaegerPort
	println(jaegerConfig)

	tracer, closer := tracing.Init("ProxyApi", jaegerConfig)
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
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("/", HomePage)
	mux.HandleFunc("/queryWordCount", QueryWordCount)
	panic(http.ListenAndServe(":3201", mux))

}

