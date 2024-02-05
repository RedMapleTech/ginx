package bind

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type TestingBody struct {
	Test string `json:"test" binding:"required"`
}

var NewTestingBody = func() interface{} { return &TestingBody{} }

func TestBindAs(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()

	e.POST("", As(NewTestingBody, WithKey("test_body")), func(ctx *gin.Context) {
		qry := ctx.MustGet("test_body").(*TestingBody)

		assert.Equal(t, "asdf", qry.Test)
		assert.Empty(t, ctx.Errors)

		ctx.Status(http.StatusOK)
	})

	body := map[string]interface{}{
		"test": "asdf",
	}
	buf, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestBindBodyTo(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()
	bodyType := TestingBody{}

	e.POST("", To(bodyType), func(ctx *gin.Context) {
		qry := ctx.MustGet("body").(*TestingBody)

		assert.Equal(t, "asdf", qry.Test)
		assert.Equal(t, "", bodyType.Test)
		assert.Empty(t, ctx.Errors)

		ctx.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/", body(map[string]interface{}{
		"test": "asdf",
	}))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestBindBodyToMultiple(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()
	bodyType := TestingBody{}

	e.POST("", To(bodyType), To(bodyType, WithKey("body2")), func(ctx *gin.Context) {
		qry := ctx.MustGet("body").(*TestingBody)
		qry2 := ctx.MustGet("body2").(*TestingBody)

		assert.Equal(t, "asdf", qry.Test)
		assert.Equal(t, "asdf", qry2.Test)
		assert.Equal(t, "", bodyType.Test)
		assert.Empty(t, ctx.Errors)

		ctx.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/", body(map[string]interface{}{
		"test": "asdf",
	}))
	req.Header.Set("Content-Type", "application/json")
	e.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestBindEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()
	bodyType := TestingBody{}

	e.POST("", To(bodyType), func(ctx *gin.Context) {
		ptr, exists := ctx.Get("body")

		assert.False(t, exists)
		assert.Nil(t, ptr)

		ctx.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestBindEmptyResponse(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()
	bodyType := TestingBody{}

	e.POST("", To(bodyType, WithResponse(true)))

	req, _ := http.NewRequest("POST", "/", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Result().StatusCode)
	assert.Equal(t, 0, len(w.Body.Bytes()))
}

func TestBindPanic(t *testing.T) {
	assert.Panics(t, func() {
		To(&TestingBody{})
	})
}

func TestBindAbort(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()

	e.POST("", func(ctx *gin.Context) {
		ctx.Next()
		assert.True(t, ctx.IsAborted())
		assert.NotEmpty(t, ctx.Errors)
	}, To(TestingBody{}, WithAbort(true)))

	req, _ := http.NewRequest("POST", "/", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestBindResponse(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()

	e.POST("", To(TestingBody{}, WithResponse(true)))

	req, _ := http.NewRequest("POST", "/", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Result().StatusCode)
}

func TestBindResponseCode(t *testing.T) {
	w := httptest.NewRecorder()
	e := gin.New()

	e.POST("", To(TestingBody{}, WithCode(http.StatusForbidden)))

	req, _ := http.NewRequest("POST", "/", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Result().StatusCode)
}

func TestBindResponseDetail(t *testing.T) {
	type validatedBody struct {
		ID string `json:"id" binding:"required,uuid4"`
	}

	w := httptest.NewRecorder()
	e := gin.New()

	e.POST("", To(validatedBody{}, WithDetail(true)))

	req, _ := http.NewRequest("POST", "/", body(map[string]string{
		"id": "not_a_uuid",
	}))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Result().StatusCode)
	assert.Equal(t, `{"code":"validation_error","errors":[{"field":"ID","rule":"uuid4"}]}`, w.Body.String())
}

func TestBindResponseDetailNotValidation(t *testing.T) {
	type validatedBody struct {
		ID string `json:"id" binding:"required,uuid4"`
	}

	w := httptest.NewRecorder()
	e := gin.New()

	e.POST("", To(validatedBody{}, WithDetail(true)))

	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"id": "invalid_json`))
	req.Header.Set("Content-Type", "application/json")

	e.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Result().StatusCode)
	assert.Equal(t, `{"code":"binding_error","error":"unexpected EOF"}`, w.Body.String())
}

func body(i interface{}) io.Reader {
	buf, _ := json.Marshal(i)
	return bytes.NewReader(buf)
}
