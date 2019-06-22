package routing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

// Binder warps the Bind method.
// This is to aproach an extensible Binder
type Binder interface {
	Bind(i interface{}, c *Context) error
}

// DefaultBinder is the dafult Binder for routing Context
type DefaultBinder struct{}

// Bind implements the `Binder#Bind` function
func (b *DefaultBinder) Bind(i interface{}, c *Context) (err error) {
	req := c.Request
	method := string(req.Header.Method())

	if req.Header.ContentLength() == 0 {
		if method == http.MethodGet || method == http.MethodDelete {
			// TODO: Implement params binding
			return NewHTTPError(fasthttp.StatusBadRequest, "GET or DELETE methods can't have body")
		}
		return NewHTTPError(fasthttp.StatusBadRequest, "Request body can't be empty")
	}

	ctype := string(req.Header.ContentType())

	switch {
	case strings.HasPrefix(ctype, MIMEApplicationJSON):
		if err = jsoniter.Unmarshal(req.Body(), i); err != nil {
			if ute, ok := err.(*json.UnmarshalTypeError); ok {
				return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset))
			} else if se, ok := err.(*json.SyntaxError); ok {
				return NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error()))
			}
			return NewHTTPError(http.StatusBadRequest, err.Error())
		}
	default:
		return NewHTTPError(fasthttp.StatusUnsupportedMediaType)
	}

	return nil
}
