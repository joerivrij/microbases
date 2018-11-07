package main

import (
	"context"
	"fmt"
	"github.com/joerivrij/microbases/shared/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

func main() {
	tracer, closer := tracing.Init("RelationalBackendApi")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	helloTo := "Relations"

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)
	defer span.Finish()

	ctx := context.Background()
	ctx = opentracing.ContextWithSpan(ctx, span)

	helloStr := formatString(ctx, helloTo)
	printHello(ctx, helloStr)
	span.Finish()

	fmt.Println("hello world")
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