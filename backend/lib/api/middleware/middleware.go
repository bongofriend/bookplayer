package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/bongofriend/bookplayer/backend/lib/config"
)

type Middleware func(http.Handler) http.Handler

type responseWriterStatusCode struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseWriterStatusCode) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.statusCode = status
}

/*
Middleware to log information about incoming requests
TODO: Add log levels for development and production
*/
func Logging(config config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rsp := &responseWriterStatusCode{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(rsp, r)
			log.Println(rsp.statusCode, r.Method, r.URL.Path, time.Since(start))
		})
	}
}

func CreateMiddlewareStack(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			m := middlewares[i]
			next = m(next)
		}
		return next
	}
}
