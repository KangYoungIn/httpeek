package handler

import (
	"net/http"

	typ "github.com/KangYoungIn/httpeek/internal/type"
	"github.com/KangYoungIn/httpeek/internal/utils"
	"github.com/gin-gonic/gin"
)

func TraceHandler(c *gin.Context) {
	var req typ.TraceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 헤더 문자열 리스트로 변환
	headers := []string{}
	for k, v := range req.Headers {
		headers = append(headers, k+": "+v)
	}

	config := typ.TraceConfig{
		Method:      req.Method,
		Headers:     headers,
		Body:        req.Body,
		ShowHeaders: req.ShowHeaders,
		ShowBody:    req.ShowBody,
	}

	if config.Method == "" {
		config.Method = "GET"
	}

	traceResult, err := utils.TraceAndCollect(req.URL, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, traceResult)
}
