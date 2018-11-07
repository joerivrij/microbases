package response

import (
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	"net/http"
	"strconv"
	openlog "github.com/opentracing/opentracing-go/log"
)

//generic method to respondwith json and log to jaeger
func RespondWithJson(w http.ResponseWriter, code int, payload interface{}, ctx context.Context) {
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
