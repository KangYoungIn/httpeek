# httpeek

`httpeek`은 HTTP 요청의 전체 네트워크 흐름(DNS, TCP, TLS, Request, Response)을 시각화하거나 JSON으로 수집할 수 있는 CLI & API 도구입니다.

- DNS → TCP → TLS → Request → Response 흐름 분석
- 인증서 정보, 응답 헤더/바디 출력
- 리다이렉트 추적 및 전체 소요 시간 측정
- CLI 또는 REST API 방식으로 사용 가능

---

## 설치

```bash
go build -o httpeek cmd/main.go
```

---

## 주요 용도

`httpeek`은 다음과 같은 상황에서 유용하게 사용할 수 있습니다.

- 운영 중인 웹 서비스, API의 네트워크 흐름 이상 유무 확인
- TLS 인증서 정보 및 만료일 검증 자동화
- DNS → TCP → TLS → 응답까지 전체 요청 단계 시각화
- 복잡한 리다이렉트 체인 확인 및 추적
- 보안 정책 적용 여부 (HSTS, CSP, 인증서 발급자 등) 확인
- API 테스트 자동화 시 응답 헤더/바디 분석

운영 환경에서 장애 조치 전 확인용 또는 디버깅/트러블슈팅 도구로도 활용할 수 있습니다.

단, 클라이언트 기준의 요청 흐름만 추적 가능하며, 백엔드 내부 로직이나 서버 간 통신 흐름은 포함되지 않습니다.


---

## 사용 예시 (CLI)

### 단일 URL 추적
```bash
./httpeek https://example.com
```

### 다중 URL 추적
```bash
./httpeek https://example.com https://github.com
```

### CLI 실행 결과 예시
```bash
$ ./httpeek --show-headers https://github.com
==================================================
분석 대상: https://github.com
==================================================
요청: GET https://github.com
-> [DNS Lookup] 20.200.245.247 (2.00ms)
-> [TCP Connect] 20.200.245.247:443 (101.00ms)
-> [TLS Handshake] CN=github.com, SANs=github.com,www.github.com, Issuer=Sectigo ECC Domain Validation Secure Server CA, Valid=2025-02-05T00:00:00Z~2026-02-05T23:59:59Z, Signature=ECDSA-SHA256, PublicKey=ECDSA (19.00ms)
-> [Request Sent]  (0.00ms)
-> [First Byte Received]  (12.00ms)
응답 코드: 200
응답 헤더:
    Server: GitHub.com
    Content-Type: text/html; charset=utf-8
    Strict-Transport-Security: max-age=31536000; includeSubdomains; preload
    ...
소요 시간: 139.00ms
```

---

## API 서버 모드

### 서버 실행
```bash
./httpeek --api
```

### POST 요청 예시
```bash
curl -X POST http://localhost:8080/trace \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://github.com/",
    "method": "GET",
    "show_headers": true,
    "show_body": true
  }'
```

### 응답 (예시)

```json
  "url": "https://github.com",
  "method": "GET",
  "resp_status": 200,{
  "url": "https://github.com",
  "method": "GET",
  "resp_status": 200,
  "resp_headers": {
    "Content-Type": ["text/html; charset=utf-8"],
    "Server": ["GitHub.com"],
    "Strict-Transport-Security": ["max-age=31536000; includeSubdomains; preload"],
    "...": ["..."]
  },
  "timeline": [
    {
      "label": "DNS Lookup",
      "start_time": "2025-03-25T00:44:34.265743+09:00",
      "duration": 18481625,
      "message": "20.200.245.247"
    },
    {
      "label": "TCP Connect",
      "start_time": "2025-03-25T00:44:34.284266+09:00",
      "duration": 7587042,
      "message": "20.200.245.247:443"
    },
    {
      "label": "TLS Handshake",
      "start_time": "2025-03-25T00:44:34.291906+09:00",
      "duration": 20276709,
      "message": "CN=github.com, SANs=github.com,www.github.com, Issuer=Sectigo ECC Domain Validation Secure Server CA, Valid=2025-02-05T00:00:00Z~2026-02-05T23:59:59Z, Signature=ECDSA-SHA256, PublicKey=ECDSA"
    },
    {
      "label": "Request Sent",
      "start_time": "2025-03-25T00:44:34.312966+09:00",
      "duration": 0
    },
    {
      "label": "First Byte Received",
      "start_time": "2025-03-25T00:44:34.324226+09:00",
      "duration": 11260167
    }
  ],
  "duration": 61096250
}
```
> `duration` 및 `timeline[*].duration` 필드는 **Go의 time.Duration(nanoseconds)** 기준의 정수 값입니다.  
> 필요 시 ms 단위로 나누어 사용하세요.
---

## 사용 라이브러리 및 라이선스

| 라이브러리              | 목적                     | 라이선스     |
|--------------------------|--------------------------|--------------|
| [cobra](https://github.com/spf13/cobra)          | CLI 파서                | Apache 2.0   |
| [gin](https://github.com/gin-gonic/gin)          | API 서버                | MIT          |
