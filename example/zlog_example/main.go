package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redmapletech/ginx/zlog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Configure zlog global settings
	zlog.SetGlobalRequestLevel(zerolog.TraceLevel)
	zlog.SetGlobalResponseLevel(zerolog.InfoLevel)

	// New gin engine
	e := gin.New()

	// Default to InfoLevel
	e.Use(zlog.Logger(zerolog.InfoLevel))

	e.GET("", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "This endpoint logs the response only")
	})

	e.GET("trace", zlog.SetLevel(zerolog.TraceLevel), func(ctx *gin.Context) {
		zlog.GetLogger(ctx).Trace().Msg("Trace Message")
		ctx.String(http.StatusOK, "This endpoint overrides the default level and logs a trace message")
	})

	e.NoRoute(func(ctx *gin.Context) {
		zlog.GetLogger(ctx).Warn().Msgf("No route handler for %s", ctx.Request.URL.Path)
		ctx.AbortWithStatus(http.StatusNotFound)
	})

	e.Run("localhost:9000")
}
