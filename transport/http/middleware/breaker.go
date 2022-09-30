package middleware

import (
	"github.com/sado0823/go-kitx/transport/http/response"
	"net/http"

	"github.com/sado0823/go-kitx/kit/breaker"
)

func Breaker() func(http.Handler) http.Handler {
	b := breaker.New()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := b.Allow()
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			writer := &response.WithCodeResponseWriter{ResponseWriter: w}
			defer func() {
				if writer.Code < http.StatusInternalServerError {
					b.MarkSuccess()
				} else {
					b.MarkFail()
				}
			}()

			next.ServeHTTP(writer, r)
		})
	}

}
