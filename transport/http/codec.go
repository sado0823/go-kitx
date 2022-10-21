package http

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sado0823/go-kitx/errorx"
	"github.com/sado0823/go-kitx/pkg/encoding"
)

type (
	Redirector interface {
		Redirect() (path string, code int)
	}

	Request http.Request

	ResponseWriter http.ResponseWriter

	Flusher http.Flusher

	DecodeRequestFunc func(req *http.Request, data interface{}) (err error)
	EncodeRequestFunc func(ctx context.Context, contentType string, req interface{}) (body []byte, err error)

	DecodeResponseFunc func(ctx context.Context, resp *http.Response, data interface{}) (err error)
	EncodeResponseFunc func(w http.ResponseWriter, req *http.Request, data interface{}) (err error)

	DecodeErrorFunc func(ctx context.Context, res *http.Response) error
	EncodeErrorFunc func(w http.ResponseWriter, req *http.Request, err error)
)

func RequestDecoder(r *http.Request, v interface{}) error {
	codec, ok := CodecForRequest(r, "Content-Type")
	if !ok {
		return errorx.BadRequest("CODEC", fmt.Sprintf("unregister Content-Type: %s", r.Header.Get("Content-Type")))
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return errorx.BadRequest("CODEC", err.Error())
	}
	if len(data) == 0 {
		return nil
	}
	if err = codec.Unmarshal(data, v); err != nil {
		return errorx.BadRequest("CODEC", fmt.Sprintf("body unmarshal %s", err.Error()))
	}
	return nil
}

func RequestEncoder(ctx context.Context, contentType string, req interface{}) (body []byte, err error) {
	subtype := contentSubtype(contentType)
	codec := encoding.GetCodec(subtype)
	if codec == nil {
		return nil, fmt.Errorf("invalid content type %s", contentType)
	}

	return codec.Marshal(req)
}

func ResponseDecoder(ctx context.Context, resp *http.Response, data interface{}) (err error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return CodecForResponse(resp).Unmarshal(body, data)
}

func ResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(Redirector); ok {
		url, code := rd.Redirect()
		http.Redirect(w, r, url, code)
		return nil
	}
	codec, _ := CodecForRequest(r, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", contentType(codec.Name()))
	_, err = w.Write(data)
	return err
}

func ErrorDecoder(ctx context.Context, res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err == nil {
		e := new(errorx.Error)
		if err = CodecForResponse(res).Unmarshal(data, e); err == nil {
			e.Code = int32(res.StatusCode)
			return e
		}
	}
	return errorx.Newf(res.StatusCode, errorx.UnknownReason, "").WithCause(err)
}

func ErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := errorx.FromError(err)
	codec, _ := CodecForRequest(r, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType(codec.Name()))
	w.WriteHeader(int(se.Code))
	_, _ = w.Write(body)
}

func CodecForRequest(r *http.Request, name string) (encoding.Codec, bool) {
	for _, accept := range r.Header[name] {
		codec := encoding.GetCodec(contentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return encoding.GetCodec("json"), false
}

func CodecForResponse(r *http.Response) encoding.Codec {
	codec := encoding.GetCodec(contentSubtype(r.Header.Get("Content-Type")))
	if codec != nil {
		return codec
	}
	return encoding.GetCodec("json")
}
