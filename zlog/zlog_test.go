package zlog

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLogRequest(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	buf := &bytes.Buffer{}
	log.Logger = zerolog.New(buf).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true})

	w := httptest.NewRecorder()
	e := gin.New()

	e.GET("", Logger(zerolog.TraceLevel))

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	e.ServeHTTP(w, req)

	fmt.Println(buf)
	lines := strings.Split(buf.String(), "\n")
	id := w.Header().Get("X-Request-ID")

	assert.NotEmpty(t, id)

	assert.Equal(t, 3, len(lines))
	assert.Contains(t, lines[0], "TRC REQ GET /")
	assert.Contains(t, lines[0], id)
	assert.Contains(t, lines[1], "DBG RES GET /")
	assert.Contains(t, lines[1], id)
}

func TestLogExtra(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	buf := &bytes.Buffer{}
	log.Logger = zerolog.New(buf).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true})

	w := httptest.NewRecorder()
	e := gin.New()

	e.GET("", Logger(zerolog.WarnLevel), func(ctx *gin.Context) {
		GetLogger(ctx).Warn().Msg("TEST")
	})

	req, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(w, req)

	fmt.Println(buf)
	lines := strings.Split(buf.String(), "\n")
	id := w.Header().Get("X-Request-ID")

	assert.NotEmpty(t, id)

	assert.Equal(t, 2, len(lines))
	assert.Contains(t, lines[0], "WRN TEST")
	assert.Contains(t, lines[0], id)
}

func TestLogChangeRequestLevel(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	buf := &bytes.Buffer{}
	log.Logger = zerolog.New(buf).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: buf, TimeFormat: time.RFC3339, NoColor: true})

	w := httptest.NewRecorder()
	e := gin.New()

	e.GET("", Logger(zerolog.WarnLevel), SetLevel(zerolog.DebugLevel), func(ctx *gin.Context) {
		GetLogger(ctx).Debug().Msg("TEST")
	})

	req, _ := http.NewRequest("GET", "/", nil)
	e.ServeHTTP(w, req)

	fmt.Println(buf)
	lines := strings.Split(buf.String(), "\n")
	id := w.Header().Get("X-Request-ID")

	assert.NotEmpty(t, id)

	assert.Equal(t, 3, len(lines))
	assert.Contains(t, lines[0], "DBG TEST")
	assert.Contains(t, lines[0], id)
	assert.Contains(t, lines[1], "DBG RES GET /")
	assert.Contains(t, lines[1], id)
}
