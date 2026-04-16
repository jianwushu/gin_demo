package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime"
	"strings"
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
		requestContentType := c.GetHeader("Content-Type")
		acceptHeader := c.GetHeader("Accept")
		logRequestBody := isJSONMediaType(requestContentType)
		logResponseBody := isJSONMediaType(acceptHeader)

		requestBodyLog := "omitted: non-json content-type"
		responseBodyLog := "omitted: non-json accept"
		responseBody := bytes.NewBuffer(nil)

		if logRequestBody {
			requestBody, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logger.Printf(
					"audit\n  request: method=%s path=%s query=%q client_ip=%s\n  response: status=%d latency=%s\n  body: request_read_error=%q",
					c.Request.Method,
					c.Request.URL.Path,
					c.Request.URL.RawQuery,
					c.ClientIP(),
					400,
					time.Since(startedAt),
					err.Error(),
				)
				c.AbortWithStatusJSON(400, gin.H{"code": 400, "message": "read request body failed"})
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			requestBodyLog = normalizeJSONLog(requestBody)
		}

		if logResponseBody {
			writer := &bodyLogWriter{ResponseWriter: c.Writer, body: responseBody}
			c.Writer = writer
		}

		c.Next()

		if logResponseBody {
			responseBodyLog = normalizeJSONLog(responseBody.Bytes())
		}

		logger.Printf(
			"audit\n  request: method=%s path=%s query=%q client_ip=%s content_type=%q accept=%q\n  response: status=%d latency=%s\n  body: request=%s\n        response=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.RawQuery,
			c.ClientIP(),
			requestContentType,
			acceptHeader,
			c.Writer.Status(),
			time.Since(startedAt),
			requestBodyLog,
			responseBodyLog,
		)
	}
}

func isJSONMediaType(headerValue string) bool {
	for _, part := range strings.Split(headerValue, ",") {
		mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(part))
		if err != nil {
			continue
		}

		if mediaType == "application/json" || strings.HasSuffix(mediaType, "+json") {
			return true
		}
	}

	return false
}

func normalizeJSONLog(data []byte) string {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return ""
	}

	var compacted bytes.Buffer
	if err := json.Compact(&compacted, trimmed); err != nil {
		return string(trimmed)
	}

	return compacted.String()
}
