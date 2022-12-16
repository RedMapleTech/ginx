// Zerolog middleware
//
// Adds request/response logging middleware, and adds the logger to the underlying context.
//
// Warning: zerolog.SetGlobalLevel will override all log level settings in this package.
// This should usually be left unset (Trace), and the default level specified in Logger().
package zlog

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	globalRequestLevel  = zerolog.TraceLevel
	globalResponseLevel = zerolog.DebugLevel
)

type loggerKey struct{}

// Logger middleare
func Logger(lvl zerolog.Level) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start tracking request duration
		start := time.Now()

		// Generate a random string to use as the request ID
		requestIDBytes := make([]byte, 8)
		io.ReadFull(rand.Reader, requestIDBytes)
		requestID := base64.RawURLEncoding.EncodeToString(requestIDBytes)
		c.Header("X-Request-ID", requestID)

		// Create a sublogger at the specified level to carry through the request chain
		logger := log.With().Str("id", requestID).Logger().Level(lvl)
		setLogger(c, &logger)

		// Log request start
		logger.WithLevel(globalRequestLevel).
			Str("method", c.Request.Method).
			Str("origin", c.GetHeader("Origin")).
			Str("path", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Msg(fmt.Sprintf("REQ %s %s %s", c.Request.Method, c.Request.URL.Path, c.ClientIP()))

		// Process remaining handlers
		c.Next()

		// Calculate elapsed and decide severity
		elapsed := time.Since(start)

		// Log response
		GetLogger(c).WithLevel(globalResponseLevel).
			Str("method", c.Request.Method).
			Str("request", c.Request.URL.Path).
			Str("ip", c.ClientIP()).
			Int("response", c.Writer.Status()).
			Int("bytes", c.Writer.Size()).
			Dur("time", elapsed).
			Msg(fmt.Sprintf("RES %s %s %d %s %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), elapsed, c.ClientIP()))
	}
}

// GetLogger returns the attached logger from the context, or the global logger if not set
func GetLogger(ctx context.Context) *zerolog.Logger {
	ictx := ctx
	if gctx, ok := ctx.(*gin.Context); ok {
		ictx = gctx.Request.Context()
	}
	if logger := ictx.Value(loggerKey{}); logger != nil {
		return logger.(*zerolog.Logger)
	}
	return &log.Logger
}

// SetLevel sets the log level for the logger attached to the context of the current handler
func SetLevel(lvl zerolog.Level) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := GetLogger(ctx).Level(lvl)
		setLogger(ctx, &logger)
	}
}

// WithLogger adds a logger to a context
func WithLogger(parent context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(parent, loggerKey{}, logger)
}

// SetGlobalRequestLevel sets the log level for the REQ http trace log line
func SetGlobalRequestLevel(lvl zerolog.Level) {
	globalRequestLevel = lvl
}

// SetGlobalResponseLevel sets the log level for the RES http trace log line
func SetGlobalResponseLevel(lvl zerolog.Level) {
	globalResponseLevel = lvl
}

func setLogger(c *gin.Context, logger *zerolog.Logger) {
	c.Request = c.Request.WithContext(WithLogger(c.Request.Context(), logger))
}
