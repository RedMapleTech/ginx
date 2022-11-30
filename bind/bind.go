// Bind middleware
//
// Configurable middleware wrapper around ctx.Bind.
// Allows the target struct and middleware behaviour to be defined in a single handler, and if successful attaches the result to the gin context.
//
// By default, any errors are silently ignored and the handler chain is not interrupted.
// This behaviour can be overridden globally for all requests, and locally within each handler to either:
//   - Abort request and set ctx.Error() for deferred handling in a higher level error middleware
//   - Abort request and send a 400 error
//   - Abort request and send a 400 error with specific validation error detail
package bind

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	v "github.com/go-playground/validator/v10"
)

var (
	defaultKey      = "body"
	defaultAbort    = false
	defaultResponse = false
	defaultDetail   = false
	defaultCode     = http.StatusBadRequest
)

type bindOpts struct {
	key      string // Context key
	abort    bool   // Abort request on a bind error and set ctx.Error
	response bool   // Abort request and send response immediately
	detail   bool   // Send validation error detail as JSON in response
	code     int    // HTTP status code if sending a response
}

// Modifier function for customising bind handler behaviour
type BindOpts func(*bindOpts) *bindOpts

// As binds the body or query into an instance of the struct pointer returned by pv
// Use To() instead unless the struct requires default values to be set
func As(pv func() interface{}, opts ...BindOpts) gin.HandlerFunc {
	bo := getBindOpts(opts...)

	return func(ctx *gin.Context) {
		v := pv()
		bindHandler(ctx, v, bo)
	}
}

// To binds the body or query into a new instance of the provided struct
func To(s interface{}, opts ...BindOpts) gin.HandlerFunc {
	bo := getBindOpts(opts...)

	t := reflect.TypeOf(s)

	// Expect a pointer to a struct of which to create a new instance
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("BindTo() must be given a struct, received %s", t.Kind()))
	}

	return func(ctx *gin.Context) {
		// New instance of pointer target struct
		v := reflect.New(t).Interface()
		bindHandler(ctx, v, bo)
	}
}

func bindHandler(ctx *gin.Context, target interface{}, opts *bindOpts) {
	var body []byte

	// Check if request has a body
	if ctx.Request.Body != nil {
		var err error

		// Read body into buffer
		body, err = io.ReadAll(ctx.Request.Body)
		if err != nil {
			return
		}

		// Replace EOF'd request body with a copy
		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	}

	// Handle binding
	if err := ctx.ShouldBind(target); err != nil {
		if opts.abort {
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			ctx.Abort()
			ctx.Error(err)
			return
		} else if opts.response {
			if vErr := (v.ValidationErrors{}); errors.As(err, &vErr) && opts.detail {
				errs := []gin.H{}
				for _, fe := range vErr {
					errs = append(errs, gin.H{
						"field": fe.Field(),
						"rule":  fe.Tag(),
					})
				}
				res := gin.H{
					"code":   "validation_error",
					"errors": errs,
				}
				ctx.AbortWithStatusJSON(opts.code, res)
			} else {
				// Error other than validation error
				ctx.AbortWithStatus(opts.code)
			}
			return
		} else {
			// Neither abort nor response requested, silently ignore
			return
		}
	}

	// Restore body and continue
	if ctx.Request.Body != nil {
		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	}
	ctx.Set(opts.key, target)
}

func getBindOpts(opts ...BindOpts) *bindOpts {
	bo := defaultBindOpts()
	for _, f := range opts {
		bo = f(bo)
	}
	return bo
}

func defaultBindOpts() *bindOpts {
	return &bindOpts{
		key:      defaultKey,
		abort:    defaultAbort,
		response: defaultResponse,
		detail:   defaultDetail,
		code:     defaultCode,
	}
}

// WithKey sets the context key to attach the unmarshalled struct for the current handler
func WithKey(key string) BindOpts {
	return func(bo *bindOpts) *bindOpts {
		bo.key = key
		return bo
	}
}

// SetDefaultKey sets the default context key to attach the unmarshalled struct for all handlers
func SetDefaultKey(key string) {
	defaultKey = key
}

// WithAbort sets the middleware to abort on error for the current handler
func WithAbort() BindOpts {
	return func(bo *bindOpts) *bindOpts {
		bo.abort = true
		return bo
	}
}

// SetDefaultAbort sets the default abort on error behaviour for all handlers
func SetDefaultAbort(abort bool) {
	defaultAbort = abort
}

// WithResponse sets the middleware to send error responses for the current handler
func WithResponse() BindOpts {
	return func(bo *bindOpts) *bindOpts {
		bo.response = true
		return bo
	}
}

// SetDefaultResponse sets the default error response behaviour for all handlers
func SetDefaultResponse(response bool) {
	defaultResponse = response
}

// WithResponseCode sets the middleware to send error responses with a custom status code for the current handler
func WithResponseCode(code int) BindOpts {
	return func(bo *bindOpts) *bindOpts {
		bo.response = true
		bo.code = code
		return bo
	}
}

// SetDefaultCode sets the default error response status code for all handlers
func SetDefaultCode(code int) {
	defaultCode = code
}

// WithResponseDetail sets the middleware to send error responses with detail field for the current handler
func WithResponseDetail() BindOpts {
	return func(bo *bindOpts) *bindOpts {
		bo.response = true
		bo.detail = true
		return bo
	}
}

// SetDefaultDetail sets the default error response detail field behaviour for all handlers
func SetDefaultDetail(detail bool) {
	defaultDetail = detail
}
