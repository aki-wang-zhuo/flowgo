/**
 * HTTP 客户端调试日志：实际请求 / 原始响应格式化为可读 JSON 文本。
 */
package action

import (
	"encoding/json"
	"net/http"
)

// formatHTTPRequestDebug 序列化即将发出的请求（method / url / headers / body）。
func formatHTTPRequestDebug(req *http.Request, body string) string {
	headers := map[string]string{}
	if req != nil {
		for k, vals := range req.Header {
			if len(vals) == 0 {
				continue
			}
			headers[k] = vals[0]
			if len(vals) > 1 {
				headers[k] = vals[0] + ";…"
			}
		}
	}
	payload := map[string]interface{}{
		"method":  "",
		"url":     "",
		"headers": headers,
		"body":    body,
	}
	if req != nil {
		payload["method"] = req.Method
		if req.URL != nil {
			payload["url"] = req.URL.String()
		}
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return body
	}
	return string(b)
}

// formatHTTPResponseDebug 序列化原始响应（status / headers / body）。
func formatHTTPResponseDebug(resp *http.Response, body string) string {
	headers := map[string]string{}
	status := 0
	if resp != nil {
		status = resp.StatusCode
		for k, vals := range resp.Header {
			if len(vals) == 0 {
				continue
			}
			headers[k] = vals[0]
			if len(vals) > 1 {
				headers[k] = vals[0] + ";…"
			}
		}
	}
	payload := map[string]interface{}{
		"status":  status,
		"headers": headers,
		"body":    body,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return body
	}
	return string(b)
}
