package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WithCodeResponseWriter(t *testing.T) {

	code := http.StatusServiceUnavailable
	content := []byte(`Test_WithCodeResponseWriter`)

	t.Run("GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			codeW := &WithCodeResponseWriter{ResponseWriter: w}

			codeW.Header().Set("x-test-get", "test-get")
			codeW.WriteHeader(code)
			assert.Equal(t, code, codeW.Code)
			_, err := codeW.Write(content)
			assert.Nil(t, err)
		})

		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)

		assert.Equal(t, "test-get", resp.Header().Get("x-test-get"))
		assert.Equal(t, code, resp.Code)
		assert.Equal(t, string(content), resp.Body.String())
	})

	t.Run("POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://localhost", nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			codeW := &WithCodeResponseWriter{ResponseWriter: w}

			codeW.Header().Set("x-test-post", "test-post")
			codeW.WriteHeader(code)
			assert.Equal(t, code, codeW.Code)
			_, err := codeW.Write(content)
			assert.Nil(t, err)
		})

		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)

		assert.Equal(t, "test-post", resp.Header().Get("x-test-post"))
		assert.Equal(t, code, resp.Code)
		assert.Equal(t, string(content), resp.Body.String())
	})
}
