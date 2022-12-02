package gapi

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(
	ctx context.Context,
	req interface{},
	infor *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)
	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	logger.
		Str("method", infor.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status_text", statusCode.String()).
		Dur("duration", duration).
		Str("protocol", "grpc").
		Msg("received a gRPC request")
	return result, err
}

type ResponseRecoder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (rec *ResponseRecoder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(rec.StatusCode)
}

func (rec *ResponseRecoder) Write(body []byte) (int, error) {
	rec.Body = body
	return rec.ResponseWriter.Write(body)
}

func HTTPLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecoder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		handler.ServeHTTP(w, r)
		duration := time.Since(startTime)

		logger := log.Info()
		if rec.StatusCode != http.StatusOK {
			logger = log.Error()
		}

		logger.Str("protocol", "http").
			Str("method", r.Method).
			Str("path", r.RequestURI).
			Int("status_code", rec.StatusCode).
			Str("status_text", http.StatusText(rec.StatusCode)).
			Dur("duration", duration).
			Msg("received a HTTP request")
	})
}
