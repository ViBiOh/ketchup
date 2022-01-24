# ketchup

[![Build](https://github.com/ViBiOh/ketchup/workflows/Build/badge.svg)](https://github.com/ViBiOh/ketchup/actions)
[![codecov](https://codecov.io/gh/ViBiOh/ketchup/branch/main/graph/badge.svg)](https://codecov.io/gh/ViBiOh/ketchup)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=ViBiOh_ketchup&metric=alert_status)](https://sonarcloud.io/dashboard?id=ViBiOh_ketchup)

Thanks to [OpenEmoji](https://openmoji.org) for favicon.

Thanks to [FontAwesome](https://fontawesome.com) for icons.

> Check your GitHub, Helm, Docker, NPM or Pypi dependencies every day or week at 8am and send a digest by email.

![](ketchup.png)

I do it mostly for myself, but if you want to support me, you can [!["Buy Me A Tea"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/vibioh)

## Features

- Timezone aware cron
- Simple and plain HTML interface, mobile ready
- Healthcheck available on `/health` endpoint
- Version available on `/version` endpoint
- Prometheus metrics for Golang and HTTP available on `/metrics` endpoint
- Graceful shutdown of HTTP listener
- Docker/Kubernetes ready container
- [12factor app](https://12factor.net) compliant

## Getting started

### Prerequisites

You need a [GitHub OAuth Token](https://github.com/settings/tokens), with no particular permission, for having a decent rate-limiting when querying GitHub. Configuration is done by passing `-githubToken` arg or setting equivalent environment variable (cf. [Usage](#usage) section)

You need a Postgres database for storing your datas. I personnaly use free tier of [ElephantSQL](https://www.elephantsql.com). Once setup, you _have to_ to create schema with [Auth DDL](https://github.com/ViBiOh/auth/blob/main/ddl.sql) and [Ketchup DDL](sql/ddl.sql). Configuration is done by passing `-dbHost`, `-dbName`, `-dbUser`, `-dbPass` args or setting equivalent environment variables (cf. [Usage](#usage) section).

You need a Redis instance for storing captcha token and distributed locks accross multiples instances. Configuration is done by passing `-redisAddress`, `-redisPassword`, `-redisDatabase` args or setting equivalent environment variables (cf. [Usage](#usage) section).

In order to send email, you must configure a [mailer](https://github.com/ViBiOh/mailer#getting-started). Configuration is done by passing `-mailerURL` arg or setting equivalent environment variable (cf. [Usage](#usage) section).

### Installation

Golang binary is built with static link. You can download it directly from the [GitHub Release page](https://github.com/ViBiOh/ketchup/releases) or build it by yourself by cloning this repo and running `make`.

A Docker image is available for `amd64`, `arm` and `arm64` platforms on Docker Hub: [vibioh/ketchup](https://hub.docker.com/r/vibioh/ketchup/tags).

You can configure app by passing CLI args or environment variables (cf. [Usage](#usage) section). CLI override environment variables.

You'll find a Kubernetes exemple in the [`infra/`](infra/) folder, using my [`app chart`](https://github.com/ViBiOh/charts/tree/main/app)

## Endpoints

- `GET /health`: healthcheck of server, always respond [`okStatus (default 204)`](#usage)
- `GET /ready`: checks external dependencies availability and then respond [`okStatus (default 204)`](#usage) or `503` during [`graceDuration`](#usage) when `SIGTERM` is received
- `GET /version`: value of `VERSION` environment variable
- `GET /metrics`: Prometheus metrics, on a dedicated port [`prometheusPort (default 9090)`](#usage)

## Usage

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of ketchup:
  -address string
        [server] Listen address {KETCHUP_ADDRESS}
  -cert string
        [server] Certificate file {KETCHUP_CERT}
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
        [owasp] Content-Security-Policy {KETCHUP_CSP} (default "default-src 'self'; base-uri 'self'; script-src 'self' 'httputils-nonce'; style-src 'self' 'httputils-nonce'")
  -dbHost string
        [db] Host {KETCHUP_DB_HOST}
  -dbMaxConn uint
        [db] Max Open Connections {KETCHUP_DB_MAX_CONN} (default 5)
  -dbName string
        [db] Name {KETCHUP_DB_NAME}
  -dbPass string
        [db] Pass {KETCHUP_DB_PASS}
  -dbPort uint
        [db] Port {KETCHUP_DB_PORT} (default 5432)
  -dbSslmode string
        [db] SSL Mode {KETCHUP_DB_SSLMODE} (default "disable")
  -dbTimeout uint
        [db] Connect timeout {KETCHUP_DB_TIMEOUT} (default 10)
  -dbUser string
        [db] User {KETCHUP_DB_USER}
  -dockerPassword string
        [docker] Registry Password {KETCHUP_DOCKER_PASSWORD}
  -dockerUsername string
        [docker] Registry Username {KETCHUP_DOCKER_USERNAME}
  -frameOptions string
        [owasp] X-Frame-Options {KETCHUP_FRAME_OPTIONS} (default "deny")
  -githubToken string
        [github] OAuth Token {KETCHUP_GITHUB_TOKEN}
  -graceDuration string
        [http] Grace duration when SIGTERM received {KETCHUP_GRACE_DURATION} (default "30s")
  -hsts
        [owasp] Indicate Strict Transport Security {KETCHUP_HSTS} (default true)
  -idleTimeout string
        [server] Idle Timeout {KETCHUP_IDLE_TIMEOUT} (default "2m")
  -key string
        [server] Key file {KETCHUP_KEY}
  -loggerJson
        [logger] Log format as JSON {KETCHUP_LOGGER_JSON}
  -loggerLevel string
        [logger] Logger level {KETCHUP_LOGGER_LEVEL} (default "INFO")
  -loggerLevelKey string
        [logger] Key for level in JSON {KETCHUP_LOGGER_LEVEL_KEY} (default "level")
  -loggerMessageKey string
        [logger] Key for message in JSON {KETCHUP_LOGGER_MESSAGE_KEY} (default "message")
  -loggerTimeKey string
        [logger] Key for timestamp in JSON {KETCHUP_LOGGER_TIME_KEY} (default "time")
  -mailerName string
        [mailer] HTTP Username or AMQP Exchange name {KETCHUP_MAILER_NAME} (default "mailer")
  -mailerPassword string
        [mailer] HTTP Pass {KETCHUP_MAILER_PASSWORD}
  -mailerURL string
        [mailer] URL (https?:// or amqps?://) {KETCHUP_MAILER_URL}
  -minify
        Minify HTML {KETCHUP_MINIFY} (default true)
  -notifierPushUrl string
        [notifier] Pushgateway URL {KETCHUP_NOTIFIER_PUSH_URL}
  -okStatus int
        [http] Healthy HTTP Status code {KETCHUP_OK_STATUS} (default 204)
  -pathPrefix string
        Root Path Prefix {KETCHUP_PATH_PREFIX}
  -port uint
        [server] Listen port (0 to disable) {KETCHUP_PORT} (default 1080)
  -prometheusAddress string
        [prometheus] Listen address {KETCHUP_PROMETHEUS_ADDRESS}
  -prometheusCert string
        [prometheus] Certificate file {KETCHUP_PROMETHEUS_CERT}
  -prometheusGzip
        [prometheus] Enable gzip compression of metrics output {KETCHUP_PROMETHEUS_GZIP}
  -prometheusIdleTimeout string
        [prometheus] Idle Timeout {KETCHUP_PROMETHEUS_IDLE_TIMEOUT} (default "10s")
  -prometheusIgnore string
        [prometheus] Ignored path prefixes for metrics, comma separated {KETCHUP_PROMETHEUS_IGNORE}
  -prometheusKey string
        [prometheus] Key file {KETCHUP_PROMETHEUS_KEY}
  -prometheusPort uint
        [prometheus] Listen port (0 to disable) {KETCHUP_PROMETHEUS_PORT} (default 9090)
  -prometheusReadTimeout string
        [prometheus] Read Timeout {KETCHUP_PROMETHEUS_READ_TIMEOUT} (default "5s")
  -prometheusShutdownTimeout string
        [prometheus] Shutdown Timeout {KETCHUP_PROMETHEUS_SHUTDOWN_TIMEOUT} (default "5s")
  -prometheusWriteTimeout string
        [prometheus] Write Timeout {KETCHUP_PROMETHEUS_WRITE_TIMEOUT} (default "10s")
  -publicURL string
        Public URL {KETCHUP_PUBLIC_URL} (default "https://ketchup.vibioh.fr")
  -readTimeout string
        [server] Read Timeout {KETCHUP_READ_TIMEOUT} (default "5s")
  -redisAddress string
        [redis] Redis Address {KETCHUP_REDIS_ADDRESS} (default "localhost:6379")
  -redisAlias string
        [redis] Connection alias, for metric {KETCHUP_REDIS_ALIAS}
  -redisDatabase int
        [redis] Redis Database {KETCHUP_REDIS_DATABASE}
  -redisPassword string
        [redis] Redis Password, if any {KETCHUP_REDIS_PASSWORD}
  -schedulerEnabled
        [scheduler] Enable cron job {KETCHUP_SCHEDULER_ENABLED} (default true)
  -schedulerHour string
        [scheduler] Hour of cron, 24-hour format {KETCHUP_SCHEDULER_HOUR} (default "08:00")
  -schedulerTimezone string
        [scheduler] Timezone {KETCHUP_SCHEDULER_TIMEZONE} (default "Europe/Paris")
  -shutdownTimeout string
        [server] Shutdown Timeout {KETCHUP_SHUTDOWN_TIMEOUT} (default "10s")
  -title string
        Application title {KETCHUP_TITLE} (default "Ketchup")
  -url string
        [alcotest] URL to check {KETCHUP_URL}
  -userAgent string
        [alcotest] User-Agent for check {KETCHUP_USER_AGENT} (default "Alcotest")
  -writeTimeout string
        [server] Write Timeout {KETCHUP_WRITE_TIMEOUT} (default "10s")
```

## Contributing

Thanks for your interest in contributing! There are many ways to contribute to this project. [Get started here](CONTRIBUTING.md).

## CI

Following variables are required for CI:

|            Name            |           Purpose           |
| :------------------------: | :-------------------------: |
|      **DOCKER_USER**       | for publishing Docker image |
|      **DOCKER_PASS**       | for publishing Docker image |
| **SCRIPTS_NO_INTERACTIVE** | for disabling prompt in CI  |
