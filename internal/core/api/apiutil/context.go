package apiutil

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/go-playground/form"
)

type ContextKey int

const (
	CtxKey ContextKey = iota
)

// FromRequest gets the underlying Context from r.Context()
func FromRequest(w http.ResponseWriter, r *http.Request) Context {
	var ctx Context
	var ok bool
	ctx, ok = r.Context().Value(CtxKey).(*contextImpl)
	if !ok {
		ctx = NewContext(w, r)
	}
	return ctx
}

// Context represents a request-specific wrapper around an HTTP request and response.
// It exposes convenience methods for request binding, response writing, and request-scoped storage.
type Context interface {
	ContextBinder
	ContextWriter

	// Header returns the response header map for the current request.
	Header() http.Header

	// Writer returns the underlying HTTP response writer.
	Writer() http.ResponseWriter

	// Request returns the underlying HTTP request.
	Request() *http.Request

	// SetRequest overrides the underlying http request. Used on context injection.
	SetRequest(r *http.Request)

	// Set stores a value in the request context under the provided key.
	Set(key any, val any)

	// Get returns a value from the request context by key.
	Get(key any) any

	// PathValue returns the value for the named path wildcard in the ServeMux pattern that matched the request.
	PathValue(key string) string

	// QueryParam returns the first value associated with the given key.
	QueryParam(key string) string
}

// ContextBinder defines request body binding behavior for a context.
type ContextBinder interface {
	// BindJSON decodes the request body as JSON into the provided destination.
	BindJSON(dst any) error

	// BindQueryParams decodes the request url query params into the provided destination.
	BindQueryParams(dst any) error

	// BindHeaders decodes the request headers into the provided destination.
	BindHeaders(dst any) error

	// BindPathValues decodes the request path values into the provided destination.
	BindPathValues(dst any) error
}

// ContextWriter defines common response writing operations for a context.
type ContextWriter interface {
	// Status writes only an HTTP status code to the response.
	Status(status int) error

	// Bytes sends a raw byte payload with the given status and content headers.
	Bytes(status int, blob []byte) error

	// String sends a plain text response formatted with the given arguments.
	String(status int, format string, args ...any) error

	// JSONBlob sends a raw JSON byte payload with the specified status.
	JSONBlob(status int, blob []byte) error

	// JSON serializes a value as JSON and writes it to the response.
	JSON(status int, v any) error

	// File serves a file from disk for the current request.
	File(filepath string)

	// Error writes a generic JSON error object with the given status code.
	Error(status int, msg string) error

	// Redirect sends an HTTP redirect to the specified URL with the given status code.
	Redirect(status int, url string) error
}

// contextImpl is the concrete implementation of Context used internally by handlers.
type contextImpl struct {
	rw http.ResponseWriter

	req *http.Request

	formDecoder *form.Decoder
}

var _ Context = (*contextImpl)(nil)

// NewContext creates a new API utility context that wraps an HTTP response writer
// and request for use in handlers.
func NewContext(writer http.ResponseWriter, request *http.Request) Context {
	ctx := new(contextImpl)
	ctx.rw = writer
	ctx.req = request
	ctx.formDecoder = form.NewDecoder()
	return ctx
}

func (c *contextImpl) Header() http.Header {
	return c.rw.Header()
}

func (c *contextImpl) Writer() http.ResponseWriter {
	return c.rw
}

func (c *contextImpl) Request() *http.Request {
	return c.req
}

func (c *contextImpl) SetRequest(r *http.Request) {
	c.req = r
}

func (c *contextImpl) Set(key any, val any) {
	ctx := context.WithValue(c.req.Context(), key, val)
	c.req = c.req.WithContext(ctx)
}

func (c *contextImpl) Get(key any) any {
	return c.req.Context().Value(key)
}

func (c *contextImpl) PathValue(key string) string {
	return c.req.PathValue(key)
}

func (c *contextImpl) QueryParam(key string) string {
	return c.req.URL.Query().Get(key)
}

/**** WRITERS ****/

func (c *contextImpl) Status(status int) error {
	c.rw.WriteHeader(status)
	return nil
}

func (c *contextImpl) Bytes(status int, blob []byte) error {
	c.Header().Set("Content-Type", "application/octet-stream")
	c.Header().Set("Content-Length", strconv.Itoa(len(blob)))

	c.rw.WriteHeader(status)
	_, err := c.rw.Write(blob)
	return err
}

func (c *contextImpl) String(status int, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)

	c.Header().Set("Content-Type", "text/plain")
	c.Header().Set("Content-Length", strconv.Itoa(len(msg)))

	c.rw.WriteHeader(status)
	_, err := c.rw.Write([]byte(msg))
	return err
}

func (c *contextImpl) JSONBlob(status int, blob []byte) error {
	c.Header().Set("Content-Type", "application/json")
	c.Header().Set("Content-Length", strconv.Itoa(len(blob)))

	c.rw.WriteHeader(status)
	_, err := c.rw.Write(blob)
	return err
}

func (c *contextImpl) JSON(status int, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(c.rw, err.Error(), http.StatusInternalServerError)
		return err
	}
	return c.JSONBlob(status, b)
}

func (c *contextImpl) File(filepath string) {
	http.ServeFile(c.rw, c.req, filepath)
}

func (c *contextImpl) Error(status int, msg string) error {
	resErr := struct {
		Error string `json:"error"`
	}{
		Error: msg,
	}
	return c.JSON(status, resErr)
}

func (c *contextImpl) Redirect(status int, url string) error {
	http.Redirect(c.rw, c.req, url, status)
	return nil
}

/**** Binders ****/

func (c *contextImpl) BindJSON(v any) error {
	return json.NewDecoder(c.req.Body).Decode(v)
}

func (c *contextImpl) BindQueryParams(v any) error {
	values := c.req.URL.Query()
	return c.decodeMap(v, values)
}

func (c *contextImpl) BindHeaders(v any) error {
	values := c.req.Header
	return c.decodeMap(v, url.Values(values))
}

func (c *contextImpl) BindPathValues(v any) error {
	values := make(url.Values)

	val := reflect.ValueOf(v).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("form")
		if tag == "" {
			tag = field.Name
		}

		if pathVal := c.req.PathValue(tag); pathVal != "" {
			values.Set(tag, pathVal)
		}
	}

	return c.decodeMap(v, values)
}

func (c *contextImpl) decodeMap(dest any, source url.Values) error {
	return c.formDecoder.Decode(dest, source)
}
