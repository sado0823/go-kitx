package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBreaker_Handle(t *testing.T) {

	t.Run("accept", func(t *testing.T) {
		breakerNext := Breaker()
		mux := breakerNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("x-test", "x-test")
			_, err := w.Write([]byte("test"))
			assert.Nil(t, err)
		}))

		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "x-test", resp.Header().Get("x-test"))
		assert.Equal(t, "test", resp.Body.String())
	})

	t.Run("reject", func(t *testing.T) {
		breakerNext := Breaker()
		mux := breakerNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusGatewayTimeout)
		}))

		req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusGatewayTimeout, resp.Code)
	})

}

func TestBreaker_Use(t *testing.T) {
	t.Run("4xx", func(t *testing.T) {
		breakerNext := Breaker()
		mux := breakerNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))

		for i := 0; i < 1000; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			resp := httptest.NewRecorder()
			mux.ServeHTTP(resp, req)
		}

		tries := 0
		for i := 0; i < 100; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			resp := httptest.NewRecorder()
			mux.ServeHTTP(resp, req)
			if resp.Code == http.StatusBadRequest {
				tries++
			}
		}
		assert.EqualValues(t, 100, tries)
	})

	t.Run("5xx", func(t *testing.T) {
		breakerNext := Breaker()
		mux5xx := breakerNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))

		for i := 0; i < 1000; i++ {
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			resp := httptest.NewRecorder()
			mux5xx.ServeHTTP(resp, req)
		}

		drops := 0
		for i := 0; i < 100; i++ {
			if i%2 == 0 || i%5 == 0 {
				time.Sleep(time.Millisecond * 20)
			}
			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			resp := httptest.NewRecorder()
			mux5xx.ServeHTTP(resp, req)
			if resp.Code == http.StatusServiceUnavailable {
				drops++
			}
		}
		// total - (1.5*accept+5) / (total + 1)
		// dropRatio = (1100 -5)/ (1100 + 1) = 0.995
		fmt.Println("drops:----", drops)
		assert.Equal(t, true, drops >= 85)
	})
}
