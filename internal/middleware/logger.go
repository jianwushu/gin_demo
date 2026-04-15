package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func Audit(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()

		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Printf(
				"audit method=%s path=%s query=%s status=%d latency=%s client_ip=%s request_body_read_error=%q",
				c.Request.Method,
				c.Request.URL.Path,
				c.Request.URL.RawQuery,
				400,
				time.Since(startedAt),
				c.ClientIP(),
				err.Error(),
			)
			c.AbortWithStatusJSON(400, gin.H{"code": 400, "message": "read request body failed"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		responseBody := bytes.NewBuffer(nil)
		writer := &bodyLogWriter{ResponseWriter: c.Writer, body: responseBody}
		c.Writer = writer

		c.Next()

		logger.Printf(
			"audit method=%s path=%s query=%s status=%d latency=%s client_ip=%s request_body=%s response_body=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
			c.Writer.Status(),
			time.Since(startedAt),
			c.ClientIP(),
			fmt.Sprintf("%q", string(requestBody)),
			fmt.Sprintf("%q", responseBody.String()),
		)
	}
}
