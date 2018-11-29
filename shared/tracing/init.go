package tracing

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
)

// Init returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func Init(service string, host string) (opentracing.Tracer, io.Closer) {
//	cfg, err := config.FromEnv()
//	if err != nil {
//		// parsing errors might happen here, such as when we get a string where we expect a number
//		log.Printf("Could not parse Jaeger env vars: %s", err.Error())
//	}
	println(host)
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
			//LocalAgentHostPort: host,
		},
	}


	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func PrintServerInfo(ctx context.Context, serverInfo string){
	span, _ := opentracing.StartSpanFromContext(ctx, "ServerInfo")
	defer span.Finish()

	span.LogKV("event", serverInfo)
}