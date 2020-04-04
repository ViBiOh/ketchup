# ketchup

[![Build Status](https://travis-ci.com/ViBiOh/ketchup.svg?branch=master)](https://travis-ci.com/ViBiOh/ketchup)
[![codecov](https://codecov.io/gh/ViBiOh/ketchup/branch/master/graph/badge.svg)](https://codecov.io/gh/ViBiOh/ketchup)
[![Go Report Card](https://goreportcard.com/badge/github.com/ViBiOh/ketchup)](https://goreportcard.com/report/github.com/ViBiOh/ketchup)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=ViBiOh/ketchup)](https://dependabot.com)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ViBiOh_ketchup&metric=alert_status)](https://sonarcloud.io/dashboard?id=ViBiOh_ketchup)

## CI

Following variables are required for CI:

| Name | Purpose |
|:--:|:--:|
| **DOMAIN** | for setting Traefik domain for app |
| **DOCKER_USER** | for publishing Docker image |
| **DOCKER_PASS** | for publishing Docker image |
| **SCRIPTS_NO_INTERACTIVE** | for disabling prompt in CI |

## Usage

```bash
Usage of ketchup:
  -address string
        [http] Listen address {KETCHUP_ADDRESS}
  -cert string
        [http] Certificate file {KETCHUP_CERT}
  -corsCredentials
        [cors] Access-Control-Allow-Credentials {KETCHUP_CORS_CREDENTIALS}
  -corsExpose string
        [cors] Access-Control-Expose-Headers {KETCHUP_CORS_EXPOSE}
  -corsHeaders string
        [cors] Access-Control-Allow-Headers {KETCHUP_CORS_HEADERS} (default "Content-Type")
  -corsMethods string
        [cors] Access-Control-Allow-Methods {KETCHUP_CORS_METHODS} (default "GET")
  -corsOrigin string
        [cors] Access-Control-Allow-Origin {KETCHUP_CORS_ORIGIN} (default "*")
  -csp string
        [owasp] Content-Security-Policy {KETCHUP_CSP} (default "default-src 'self'; base-uri 'self'")
  -frameOptions string
        [owasp] X-Frame-Options {KETCHUP_FRAME_OPTIONS} (default "deny")
  -githubToken string
        [github] OAuth Token {KETCHUP_GITHUB_TOKEN}
  -graceDuration string
        [http] Grace duration when SIGTERM received {KETCHUP_GRACE_DURATION} (default "15s")
  -hsts
        [owasp] Indicate Strict Transport Security {KETCHUP_HSTS} (default true)
  -key string
        [http] Key file {KETCHUP_KEY}
  -okStatus int
        [http] Healthy HTTP Status code {KETCHUP_OK_STATUS} (default 204)
  -port uint
        [http] Listen port {KETCHUP_PORT} (default 1080)
  -prometheusPath string
        [prometheus] Path for exposing metrics {KETCHUP_PROMETHEUS_PATH} (default "/metrics")
  -swaggerTitle string
        [swagger] API Title {KETCHUP_SWAGGER_TITLE} (default "API")
  -swaggerVersion string
        [swagger] API Version {KETCHUP_SWAGGER_VERSION} (default "1.0.0")
  -url string
        [alcotest] URL to check {KETCHUP_URL}
  -userAgent string
        [alcotest] User-Agent for check {KETCHUP_USER_AGENT} (default "Alcotest")
```
