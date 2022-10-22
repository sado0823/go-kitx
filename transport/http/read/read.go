package read

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

const (
	// ContentBinaryHeaderValue header value for binary data.
	ContentBinaryHeaderValue = "application/octet-stream"
	// ContentWebassemblyHeaderValue header value for web assembly files.
	ContentWebassemblyHeaderValue = "application/wasm"
	// ContentHTMLHeaderValue is the  string of text/html response header's content type value.
	ContentHTMLHeaderValue = "text/html"
	// ContentJSONHeaderValue header value for JSON data.
	ContentJSONHeaderValue = "application/json"
	// ContentJSONProblemHeaderValue header value for JSON API problem error.
	// Read more at: https://tools.ietf.org/html/rfc7807
	ContentJSONProblemHeaderValue = "application/problem+json"
	// ContentXMLProblemHeaderValue header value for XML API problem error.
	// Read more at: https://tools.ietf.org/html/rfc7807
	ContentXMLProblemHeaderValue = "application/problem+xml"
	// ContentJavascriptHeaderValue header value for JSONP & Javascript data.
	ContentJavascriptHeaderValue = "text/javascript"
	// ContentTextHeaderValue header value for Text data.
	ContentTextHeaderValue = "text/plain"
	// ContentXMLHeaderValue header value for XML data.
	ContentXMLHeaderValue = "text/xml"
	// ContentXMLUnreadableHeaderValue obselete header value for XML.
	ContentXMLUnreadableHeaderValue = "application/xml"
	// ContentMarkdownHeaderValue custom key/content type, the real is the text/html.
	ContentMarkdownHeaderValue = "text/markdown"
	// ContentYAMLHeaderValue header value for YAML data.
	ContentYAMLHeaderValue = "application/x-yaml"
	// ContentYAMLTextHeaderValue header value for YAML plain text.
	ContentYAMLTextHeaderValue = "text/yaml"
	// ContentProtobufHeaderValue header value for Protobuf messages data.
	ContentProtobufHeaderValue = "application/x-protobuf"
	// ContentMsgPackHeaderValue header value for MsgPack data.
	ContentMsgPackHeaderValue = "application/msgpack"
	// ContentMsgPack2HeaderValue alternative header value for MsgPack data.
	ContentMsgPack2HeaderValue = "application/x-msgpack"
	// ContentFormHeaderValue header value for post form data.
	ContentFormHeaderValue = "application/x-www-form-urlencoded"
	// ContentFormMultipartHeaderValue header value for post multipart form data.
	ContentFormMultipartHeaderValue = "multipart/form-data"
	// ContentMultipartRelatedHeaderValue header value for multipart related data.
	ContentMultipartRelatedHeaderValue = "multipart/related"
	// ContentGRPCHeaderValue Content-Type header value for gRPC.
	ContentGRPCHeaderValue = "application/grpc"
)

var (
	emptyFunc = func() {}
)

// GetForm returns the request form (url queries, post or multipart) values.
func GetForm(r *http.Request, postMaxMemory int64, resetBody bool) (form url.Values, found bool) {
	/*
		net/http/request.go#1219
		for k, v := range f.Value {
			r.Form[k] = append(r.Form[k], v...)
			// r.PostForm should also be populated. See Issue 9305.
			r.PostForm[k] = append(r.PostForm[k], v...)
		}
	*/

	if form := r.Form; len(form) > 0 {
		return form, true
	}

	if form := r.PostForm; len(form) > 0 {
		return form, true
	}

	if m := r.MultipartForm; m != nil {
		if len(m.Value) > 0 {
			return m.Value, true
		}
	}

	if resetBody {
		// on POST, PUT and PATCH it will read the form values from request body otherwise from URL queries.
		if m := r.Method; m == "POST" || m == "PUT" || m == "PATCH" {
			body, restoreBody, err := GetBody(r, resetBody)
			if err != nil {
				return nil, false
			}
			setBody(r, body)    // so the ctx.request.Body works
			defer restoreBody() // so the next GetForm calls work.

			// r.Body = io.NopCloser(io.TeeReader(r.Body, buf))
		}
	}

	// ParseMultipartForm calls `request.ParseForm` automatically
	// therefore we don't need to call it here, although it doesn't hurt.
	// After one call to ParseMultipartForm or ParseForm,
	// subsequent calls have no effect, are idempotent.
	err := r.ParseMultipartForm(postMaxMemory)
	// if resetBody {
	// 	r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))
	// }
	if err != nil && err != http.ErrNotMultipart {
		return nil, false
	}

	if form := r.Form; len(form) > 0 {
		return form, true
	}

	if form := r.PostForm; len(form) > 0 {
		return form, true
	}

	if m := r.MultipartForm; m != nil {
		if len(m.Value) > 0 {
			return m.Value, true
		}
	}

	return nil, false
}

// GetBody reads and returns the request body.
func GetBody(r *http.Request, resetBody bool) (data []byte, reuse func(), err error) {
	data, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	if resetBody {
		// * remember, Request.Body has no Bytes(), we have to consume them first
		// and after re-set them to the body, this is the only solution.
		return data, func() {
			setBody(r, data)
		}, nil
	}

	return data, emptyFunc, nil
}

func setBody(r *http.Request, data []byte) {
	r.Body = io.NopCloser(bytes.NewBuffer(data))
}

func GetQuery(r *http.Request) url.Values {
	return r.URL.Query()
}
