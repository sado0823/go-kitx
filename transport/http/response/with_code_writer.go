package response

import "net/http"

type WithCodeResponseWriter struct {
	Code   int
	Writer http.ResponseWriter
}

func (w *WithCodeResponseWriter) WriteHeader(code int) {
	w.Writer.WriteHeader(code)
	w.Code = code
}

func (w *WithCodeResponseWriter) Reset(res http.ResponseWriter) {
	w.Writer = res
	w.Code = http.StatusOK
}

func (w *WithCodeResponseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

func (w *WithCodeResponseWriter) Header() http.Header {
	return w.Writer.Header()
}
