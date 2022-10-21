package http

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sado0823/go-kitx/transport"
	"github.com/sado0823/go-kitx/transport/http/binding"
	"github.com/sado0823/go-kitx/transport/http/response"
	"github.com/sado0823/go-kitx/transport/pbchain"

	"github.com/gorilla/mux"
)

var (
	_ Context = (*contextx)(nil)
)

// Context is an HTTP Context.
type (
	Context interface {
		context.Context
		Vars() url.Values
		Query() url.Values
		Form() url.Values
		Header() http.Header
		Request() *http.Request
		Response() http.ResponseWriter
		Middleware(pbchain.Handler) pbchain.Handler
		Bind(interface{}) error
		BindVars(interface{}) error
		BindQuery(interface{}) error
		BindForm(interface{}) error
		Returns(interface{}, error) error
		Result(int, interface{}) error
		JSON(int, interface{}) error
		XML(int, interface{}) error
		String(int, string) error
		Blob(int, string, []byte) error
		Stream(int, string, io.Reader) error
		Reset(http.ResponseWriter, *http.Request)
	}

	contextx struct {
		router *Router
		req    *http.Request
		res    http.ResponseWriter
		w      response.DelayHeaderResponseWriter
	}
)

func (c *contextx) Header() http.Header {
	return c.req.Header
}

func (c *contextx) Vars() url.Values {
	raws := mux.Vars(c.req)
	vars := make(url.Values, len(raws))
	for k, v := range raws {
		vars[k] = []string{v}
	}
	return vars
}

func (c *contextx) Form() url.Values {
	if err := c.req.ParseForm(); err != nil {
		return url.Values{}
	}
	return c.req.Form
}

func (c *contextx) Query() url.Values {
	return c.req.URL.Query()
}

func (c *contextx) Request() *http.Request {
	return c.req
}

func (c *contextx) Response() http.ResponseWriter {
	return c.res
}

func (c *contextx) Middleware(h pbchain.Handler) pbchain.Handler {
	if tr, ok := transport.FromServerContext(c.req.Context()); ok {
		return pbchain.Chain(c.router.srv.pbchain.Match(tr.Operation())...)(h)
	}
	return pbchain.Chain(c.router.srv.pbchain.Match(c.req.URL.Path)...)(h)
}

func (c *contextx) Bind(v interface{}) error {
	return c.router.srv.reqDecoder(c.req, v)
}

func (c *contextx) BindVars(v interface{}) error {
	return binding.BindQuery(c.Vars(), v)
}

func (c *contextx) BindQuery(v interface{}) error {
	return binding.BindQuery(c.Query(), v)
}

func (c *contextx) BindForm(v interface{}) error {
	return binding.BindForm(c.req, v)
}

func (c *contextx) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	return c.router.srv.respEncoder(&c.w, c.req, v)
}

func (c *contextx) Result(code int, v interface{}) error {
	c.w.WriteHeader(code)
	return c.router.srv.respEncoder(&c.w, c.req, v)
}

func (c *contextx) JSON(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/json")
	c.res.WriteHeader(code)
	return json.NewEncoder(c.res).Encode(v)
}

func (c *contextx) XML(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/xml")
	c.res.WriteHeader(code)
	return xml.NewEncoder(c.res).Encode(v)
}

func (c *contextx) String(code int, text string) error {
	c.res.Header().Set("Content-Type", "text/plain")
	c.res.WriteHeader(code)
	_, err := c.res.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

func (c *contextx) Blob(code int, contentType string, data []byte) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(code)
	_, err := c.res.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *contextx) Stream(code int, contentType string, rd io.Reader) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(code)
	_, err := io.Copy(c.res, rd)
	return err
}

func (c *contextx) Reset(res http.ResponseWriter, req *http.Request) {
	c.w.Reset(res)
	c.res = res
	c.req = req
}

func (c *contextx) Deadline() (time.Time, bool) {
	if c.req == nil {
		return time.Time{}, false
	}
	return c.req.Context().Deadline()
}

func (c *contextx) Done() <-chan struct{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Done()
}

func (c *contextx) Err() error {
	if c.req == nil {
		return context.Canceled
	}
	return c.req.Context().Err()
}

func (c *contextx) Value(key interface{}) interface{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Value(key)
}
