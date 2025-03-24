package typ

import (
	"net/http"
	"time"
)

type TraceEvent struct {
	Label     string        `json:"label"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Message   string        `json:"message,omitempty"`
}

type TraceConfig struct {
	Method      string
	Headers     []string
	Body        string
	ShowHeaders bool
	ShowBody    bool
}

type RequestTrace struct {
	URL         string        `json:"url"`
	Method      string        `json:"method"`
	Headers     http.Header   `json:"headers"`
	ReqBody     string        `json:"req_body,omitempty"`
	RespStatus  int           `json:"resp_status"`
	RespHeaders http.Header   `json:"resp_headers"`
	RespBody    string        `json:"resp_body,omitempty"`
	Timeline    []TraceEvent  `json:"timeline"`
	RedirectTo  *RequestTrace `json:"redirect_to,omitempty"`
	Duration    time.Duration `json:"duration"`
}

type TraceRequest struct {
	URL         string            `json:"url" binding:"required"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	ShowHeaders bool              `json:"show_headers,omitempty"`
	ShowBody    bool              `json:"show_body,omitempty"`
}

type TraceResponse struct {
	Status string `json:"status"`
}
