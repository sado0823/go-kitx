package response

import "net/http"

type DelayHeaderResponseWriter struct {
	Code   int
	Writer http.ResponseWriter
}

func (w *DelayHeaderResponseWriter) WriteHeader(code int) {
	w.Code = code
}

func (w *DelayHeaderResponseWriter) Reset(res http.ResponseWriter) {
	w.Writer = res
	w.Code = http.StatusOK
}

func (w *DelayHeaderResponseWriter) Write(data []byte) (int, error) {
	w.Writer.WriteHeader(w.Code)
	return w.Writer.Write(data)
}

func (w *DelayHeaderResponseWriter) Header() http.Header {
	return w.Writer.Header()
}
