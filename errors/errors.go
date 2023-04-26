// Error wrapper
//
// Shorthand for checking error states, and conditionally aborting with specified status and JSON body.
// Inclusion of actual error message can also be enabled.
package errors

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/richjyoung/wtf"
)

var (
	detail = false

	ErrNoDetail = fmt.Errorf("error: no detail")
)

// AbortWithError is a wrapper around AbortWithStatusJSON, returning true if the error
// was not nil and therefore the request has been aborted.
// Intended to be used as follows:
//
//	if errors.AbortWithError(ctx, err, 400, "error_short_code") {
//		return
//	}
func AbortWithError(ctx *gin.Context, err error, status int, code string) bool {
	if err != nil {
		body := gin.H{"code": code}
		if detail && !errors.Is(err, ErrNoDetail) {
			body["error"] = wtf.IsThisError(err)
		}
		ctx.AbortWithStatusJSON(status, body)
		return true
	}
	return false
}

// AbortWith is shorthand for calling AbortWithError using ErrNoDetail to suppress the error detail
func AbortWith(ctx *gin.Context, status int, code string) bool {
	return AbortWithError(ctx, ErrNoDetail, status, code)
}

// SetErrorDetailOutput sets whether the internal error message is included in the JSON response
func SetErrorDetailOutput(output bool) {
	detail = output
}
