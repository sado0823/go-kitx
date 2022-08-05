package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assert "github.com/smartystreets/goconvey/convey"
)

func TestBreaker_Handle(t *testing.T) {
	assert.Convey("TestBreaker_Handle", t, func() {
		assert.Convey("accept", func() {
			breakerNext := Breaker()
			mux := breakerNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("x-test", "x-test")
				_, err := w.Write([]byte("test"))
				assert.So(err, assert.ShouldBeNil)
			}))

			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			resp := httptest.NewRecorder()
			mux.ServeHTTP(resp, req)

			assert.So(resp.Code, assert.ShouldEqual, http.StatusOK)
			assert.So(resp.Header().Get("x-test"), assert.ShouldEqual, "x-test")
			assert.So(resp.Body.String(), assert.ShouldEqual, "test")
		})

		assert.Convey("reject", func() {
			breakerNext := Breaker()
			mux := breakerNext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusGatewayTimeout)
			}))

			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
			resp := httptest.NewRecorder()
			mux.ServeHTTP(resp, req)

			assert.So(resp.Code, assert.ShouldEqual, http.StatusGatewayTimeout)
		})

	})
}

func TestBreaker_Use(t *testing.T) {
	assert.Convey("TestBreaker_Use", t, func() {
		assert.Convey("4xx", func() {
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

			assert.So(tries, assert.ShouldEqual, 100)
		})

		assert.Convey("5xx", func() {
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
			assert.So(drops >= 85, assert.ShouldBeTrue)
		})
	})

}
