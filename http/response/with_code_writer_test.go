package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	assert "github.com/smartystreets/goconvey/convey"
)

func Test_WithCodeResponseWriter(t *testing.T) {
	assert.Convey("Test_WithCodeResponseWriter", t, func() {

		code := http.StatusServiceUnavailable
		content := []byte(`Test_WithCodeResponseWriter`)

		assert.Convey("GET", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				codeW := &WithCodeResponseWriter{ResponseWriter: w}

				codeW.Header().Set("x-test-get", "test-get")
				codeW.WriteHeader(code)
				assert.So(codeW.Code, assert.ShouldEqual, code)
				_, err := codeW.Write(content)
				assert.So(err, assert.ShouldBeNil)
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			assert.So(resp.Header().Get("x-test-get"), assert.ShouldEqual, "test-get")
			assert.So(resp.Code, assert.ShouldEqual, code)
			assert.So(resp.Body.String(), assert.ShouldEqual, string(content))
		})

		assert.Convey("POST", func() {
			req := httptest.NewRequest(http.MethodPost, "http://localhost", nil)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				codeW := &WithCodeResponseWriter{ResponseWriter: w}

				codeW.Header().Set("x-test-post", "test-post")
				codeW.WriteHeader(code)
				assert.So(codeW.Code, assert.ShouldEqual, code)
				_, err := codeW.Write(content)
				assert.So(err, assert.ShouldBeNil)
			})

			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			assert.So(resp.Header().Get("x-test-post"), assert.ShouldEqual, "test-post")
			assert.So(resp.Code, assert.ShouldEqual, code)
			assert.So(resp.Body.String(), assert.ShouldEqual, string(content))
		})

	})
}
