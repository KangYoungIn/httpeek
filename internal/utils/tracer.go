package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	typ "github.com/KangYoungIn/httpeek/internal/type"
)

// CLI 용
func TraceURL(url string, config typ.TraceConfig) error {
	trace, err := TraceCore(url, config)
	if err != nil {
		return err
	}
	PrintTrace(trace, 0, config)
	return nil
}

// API 용
func TraceAndCollect(url string, config typ.TraceConfig) (*typ.RequestTrace, error) {
	return TraceCore(url, config)
}

func TraceCore(url string, config typ.TraceConfig) (*typ.RequestTrace, error) {
	trace := &typ.RequestTrace{
		URL:      url,
		Method:   config.Method,
		Headers:  http.Header{},
		Timeline: []typ.TraceEvent{},
		ReqBody:  config.Body,
	}

	var reqBody io.Reader
	if config.Body != "" {
		reqBody = bytes.NewBufferString(config.Body)
	}

	req, err := http.NewRequest(config.Method, url, reqBody)
	if err != nil {
		return nil, err
	}

	for _, h := range config.Headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			req.Header.Set(k, v)
			trace.Headers.Add(k, v)
		}
	}

	// 타이밍 수집용 변수
	var (
		startTime                                            = time.Now()
		dnsStart, connStart, tlsStart, wroteStart, firstByte time.Time
	)

	client := CreateHTTPClient()

	traceCtx := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			var msgParts []string

			if info.Err != nil {
				msgParts = append(msgParts, fmt.Sprintf("오류: %v", info.Err))
			} else {
				for _, addr := range info.Addrs {
					msgParts = append(msgParts, addr.String())
				}
			}

			if info.Coalesced {
				msgParts = append(msgParts, "(다른 요청과 병합됨)")
			}

			trace.Timeline = append(trace.Timeline, typ.TraceEvent{
				Label:     "DNS Lookup",
				StartTime: dnsStart,
				Duration:  time.Since(dnsStart),
				Message:   strings.Join(msgParts, ", "),
			})
		},
		ConnectStart: func(network, addr string) {
			connStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			trace.Timeline = append(trace.Timeline, typ.TraceEvent{
				Label:     "TCP Connect",
				StartTime: connStart,
				Duration:  time.Since(connStart),
				Message:   addr,
			})
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if len(cs.PeerCertificates) > 0 {
				cert := cs.PeerCertificates[0]

				detail := fmt.Sprintf(
					"CN=%s, SANs=%s, Issuer=%s, Valid=%s~%s, Signature=%s, PublicKey=%s",
					cert.Subject.CommonName,
					strings.Join(cert.DNSNames, ","),
					cert.Issuer.CommonName,
					cert.NotBefore.Format(time.RFC3339),
					cert.NotAfter.Format(time.RFC3339),
					cert.SignatureAlgorithm,
					cert.PublicKeyAlgorithm,
				)

				trace.Timeline = append(trace.Timeline, typ.TraceEvent{
					Label:     "TLS Handshake",
					StartTime: tlsStart,
					Duration:  time.Since(tlsStart),
					Message:   detail,
				})
			} else {
				trace.Timeline = append(trace.Timeline, typ.TraceEvent{
					Label:     "TLS Handshake",
					StartTime: tlsStart,
					Duration:  time.Since(tlsStart),
					Message:   "(no certificate)",
				})
			}
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			wroteStart = time.Now()
			trace.Timeline = append(trace.Timeline, typ.TraceEvent{
				Label:     "Request Sent",
				StartTime: wroteStart,
			})
		},
		GotFirstResponseByte: func() {
			firstByte = time.Now()
			trace.Timeline = append(trace.Timeline, typ.TraceEvent{
				Label:     "First Byte Received",
				StartTime: firstByte,
				Duration:  firstByte.Sub(wroteStart),
			})
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(context.Background(), traceCtx))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trace.RespStatus = resp.StatusCode
	trace.RespHeaders = resp.Header

	if config.ShowBody {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			trace.RespBody = string(bodyBytes)
		}
	}

	trace.Duration = time.Since(startTime)

	// 리다이렉트 처리
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := resp.Header.Get("Location")
		if loc != "" {
			redirectTrace, err := TraceCore(loc, config)
			if err == nil {
				trace.RedirectTo = redirectTrace
			}
		}
	}

	return trace, nil
}

func PrintTrace(trace *typ.RequestTrace, level int, config typ.TraceConfig) {
	indent := strings.Repeat("    ", level)
	fmt.Printf("%s요청: %s %s\n", indent, trace.Method, trace.URL)

	for _, ev := range trace.Timeline {
		fmt.Printf("%s-> [%s] %s (%.2fms)\n", indent, ev.Label, ev.Message, float64(ev.Duration.Milliseconds()))
	}

	fmt.Printf("%s응답 코드: %d\n", indent, trace.RespStatus)

	if config.ShowHeaders {
		fmt.Printf("%s응답 헤더:\n", indent)
		for k, v := range trace.RespHeaders {
			fmt.Printf("%s    %s: %s\n", indent, k, strings.Join(v, ", "))
		}
	}

	if config.ShowBody && trace.RespBody != "" {
		fmt.Printf("%s응답 Body:\n", indent)
		fmt.Println(trace.RespBody)
	}

	fmt.Printf("%s소요 시간: %.2fms\n", indent, float64(trace.Duration.Milliseconds()))

	if trace.RedirectTo != nil {
		fmt.Printf("%s리다이렉트:\n", indent)
		PrintTrace(trace.RedirectTo, level+1, config)
	}
}

func joinAddrs(addrs []net.IPAddr) string {
	strs := []string{}
	for _, a := range addrs {
		strs = append(strs, a.String())
	}
	return strings.Join(strs, ", ")
}
