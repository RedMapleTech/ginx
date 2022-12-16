package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redmapletech/ginx/bind"
)

type TestRequestBody struct {
	Message string `json:"message,omitempty"`
}

type TestRequestQuery struct {
	Offset int `form:"offset" binding:"omitempty,gte=0"`
	Limit  int `form:"limit" binding:"omitempty,gte=10,lte=100"`
}

func NewTestRequestQuery() interface{} {
	return &TestRequestQuery{
		Limit: 10,
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()

	// Use bind.As with a provider func to set default values for query string
	e.GET("", bind.As(NewTestRequestQuery, bind.WithKey("query"), bind.WithDetail(true)), func(ctx *gin.Context) {
		qry := ctx.MustGet("query").(*TestRequestQuery)
		ctx.JSON(http.StatusOK, qry)
	})

	// Use bind.To with a struct (make sure to set request Content-Type)
	e.POST("", bind.To(TestRequestBody{}, bind.WithDetail(true)), func(ctx *gin.Context) {
		qry := ctx.MustGet("body").(*TestRequestBody)
		ctx.JSON(http.StatusOK, gin.H{
			"body": qry,
		})
	})

	e.Run("localhost:9000")
}
