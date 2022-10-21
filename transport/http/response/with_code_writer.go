package response

import "net/http"

type WithCodeResponseWriter struct {
	Code int
	http.ResponseWriter
}

func (w *WithCodeResponseWriter) WriteHeader(statusCode int) {
	w.Code = statusCode
}

func (w *WithCodeResponseWriter) Reset(res http.ResponseWriter) {
	w.ResponseWriter = res
	w.Code = http.StatusOK
}

func (w *WithCodeResponseWriter) Write(data []byte) (int, error) {
	w.ResponseWriter.WriteHeader(w.Code)
	return w.ResponseWriter.Write(data)
}
