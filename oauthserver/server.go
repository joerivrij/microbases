package main

import (
	"context"
	"fmt"
	"github.com/joerivrij/microbases/shared/tracing"
	openlog "github.com/opentracing/opentracing-go/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
	"net/http"
	"os"

	"log"
)

func main() {

	jaegerUrl := os.Getenv("JAEGER_AGENT_HOST")
	jaegerPort :=  os.Getenv("JAEGER_AGENT_PORT")
	jaegerConfig := jaegerUrl + ":" + jaegerPort
	println(jaegerConfig)

	tracer, closer := tracing.Init("OauthServer", jaegerConfig)
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3240"
	}

	span := tracer.StartSpan("OAuthServer")
	span.SetTag("event", "Starting Server")
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	logValue := fmt.Sprintf("Starting server on port %s", port)
	tracing.PrintServerInfo(ctx, logValue)
	span.Finish()

	fmt.Println("hello world")

	manager := manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// client memory store
	clientStore := store.NewClientStore()
	clientStore.Set("000000", &models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost",
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		span := opentracing.GlobalTracer().StartSpan("authorizeHandler", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ctx := context.Background()
		ctx = opentracing.ContextWithSpan(ctx, span)

		span.LogFields(
			openlog.String("method", r.Method),
			openlog.String("path", r.URL.Path),
			openlog.String("host", r.Host),
		)

		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		span := opentracing.GlobalTracer().StartSpan("tokenHandler", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ctx := context.Background()
		ctx = opentracing.ContextWithSpan(ctx, span)

		span.LogFields(
			openlog.String("method", r.Method),
			openlog.String("path", r.URL.Path),
			openlog.String("host", r.Host),
		)

		srv.HandleTokenRequest(w, r)
	})

	log.Fatal(http.ListenAndServe(":" + port, nil))
}
