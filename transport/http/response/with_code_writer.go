package response

import "net/http"

type WithCodeResponseWriter struct {
	Code int
	http.ResponseWriter
}

func (w *WithCodeResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
