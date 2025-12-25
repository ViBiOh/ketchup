# ketchup

[![Build](https://github.com/ViBiOh/ketchup/workflows/Build/badge.svg)](https://github.com/ViBiOh/ketchup/actions)

Thanks to [OpenEmoji](https://openmoji.org) for favicon.

Thanks to [FontAwesome](https://fontawesome.com) for icons.

> Check your GitHub, Helm, Docker, NPM or Pypi dependencies every day or week at 8am and send a digest by email.

![](ketchup.png)

## Features

- Simple and plain HTML interface, mobile ready
- Healthcheck available on `/health` endpoint
- Version available on `/version` endpoint
- OpenTelemetry
- Graceful shutdown of HTTP listener
- Docker/Kubernetes ready container
- [12factor app](https://12factor.net) compliant

## Getting started

### Prerequisites

You need a [GitHub OAuth Token](https://github.com/settings/tokens), with no particular permission, for having a decent rate-limiting when querying GitHub. Configuration is done by passing `-githubToken` arg or setting equivalent environment variable (cf. [Usage](#usage) section).

You need a Postgres database for storing your data. Once setup, you _have to_ to create schema with [Auth DDL](https://github.com/ViBiOh/auth/blob/main/ddl.sql) and [Ketchup DDL](sql/ddl.sql). Configuration is done by passing `-dbHost`, `-dbName`, `-dbUser`, `-dbPass` args or setting equivalent environment variables (cf. [Usage](#usage) section).

You need a Redis instance for storing captcha token and distributed locks across multiples instances. Configuration is done by passing `-redisAddress`, `-redisPassword`, `-redisDatabase` args or setting equivalent environment variables (cf. [Usage](#usage) section).

In order to send email, you must configure a [mailer](https://github.com/ViBiOh/mailer#getting-started). Configuration is done by passing `-mailerURL` arg or setting equivalent environment variable (cf. [Usage](#usage) section).

### Installation

Golang binary is built with static link. You can download it directly from the [GitHub Release page](https://github.com/ViBiOh/ketchup/releases) or build it by yourself by cloning this repo and running `make`.

You can configure app by passing CLI args or environment variables (cf. [Usage](#usage) section). CLI override environment variables.

You'll find a Kubernetes example in the [`infra/`](infra) folder, using my [`app chart`](https://github.com/ViBiOh/charts/tree/main/app)

## Endpoints

- `GET /health`: healthcheck of server, always respond [`okStatus (default 204)`](#usage)
- `GET /ready`: checks external dependencies availability and then respond [`okStatus (default 204)`](#usage) or `503` during [`graceDuration`](#usage) when close signal is received
- `GET /version`: value of `VERSION` environment variable

## Usage

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of ketchup:
  --address           string        [server] Listen address ${KETCHUP_ADDRESS}
  --capSecretKey      string        [cap] Secret Key ${KETCHUP_CAP_SECRET_KEY}
  --capSiteKey        string        [cap] Site Key ${KETCHUP_CAP_SITE_KEY}
  --capURL            string        [cap] Instance URL ${KETCHUP_CAP_URL} (default "http://cap")
  --cert              string        [server] Certificate file ${KETCHUP_CERT}
  --corsCredentials                 [cors] Access-Control-Allow-Credentials ${KETCHUP_CORS_CREDENTIALS} (default false)
  --corsExpose        string        [cors] Access-Control-Expose-Headers ${KETCHUP_CORS_EXPOSE}
  --corsHeaders       string        [cors] Access-Control-Allow-Headers ${KETCHUP_CORS_HEADERS} (default "Content-Type")
  --corsMethods       string        [cors] Access-Control-Allow-Methods ${KETCHUP_CORS_METHODS} (default "GET")
  --corsOrigin        string        [cors] Access-Control-Allow-Origin ${KETCHUP_CORS_ORIGIN} (default "*")
  --csp               string        [owasp] Content-Security-Policy ${KETCHUP_CSP} (default "default-src 'self'; base-uri 'self'; script-src 'self' 'httputils-nonce'; style-src 'self' 'httputils-nonce'")
  --dbHost            string        [db] Host ${KETCHUP_DB_HOST}
  --dbMaxConn         uint          [db] Max Open Connections ${KETCHUP_DB_MAX_CONN} (default 5)
  --dbMinConn         uint          [db] Min Open Connections ${KETCHUP_DB_MIN_CONN} (default 2)
  --dbName            string        [db] Name ${KETCHUP_DB_NAME}
  --dbPass            string        [db] Pass ${KETCHUP_DB_PASS}
  --dbPort            uint          [db] Port ${KETCHUP_DB_PORT} (default 5432)
  --dbSslmode         string        [db] SSL Mode ${KETCHUP_DB_SSLMODE} (default "disable")
  --dbUser            string        [db] User ${KETCHUP_DB_USER}
  --dockerPassword    string        [docker] Registry Password ${KETCHUP_DOCKER_PASSWORD}
  --dockerUsername    string        [docker] Registry Username ${KETCHUP_DOCKER_USERNAME}
  --extension         string        Go Template Extension ${KETCHUP_EXTENSION} (default "tmpl")
  --frameOptions      string        [owasp] X-Frame-Options ${KETCHUP_FRAME_OPTIONS} (default "deny")
  --githubToken       string        [github] OAuth Token ${KETCHUP_GITHUB_TOKEN}
  --graceDuration     duration      [http] Grace duration when signal received ${KETCHUP_GRACE_DURATION} (default 30s)
  --hsts                            [owasp] Indicate Strict Transport Security ${KETCHUP_HSTS} (default true)
  --idleTimeout       duration      [server] Idle Timeout ${KETCHUP_IDLE_TIMEOUT} (default 2m0s)
  --key               string        [server] Key file ${KETCHUP_KEY}
  --loggerJson                      [logger] Log format as JSON ${KETCHUP_LOGGER_JSON} (default false)
  --loggerLevel       string        [logger] Logger level ${KETCHUP_LOGGER_LEVEL} (default "INFO")
  --loggerLevelKey    string        [logger] Key for level in JSON ${KETCHUP_LOGGER_LEVEL_KEY} (default "level")
  --loggerMessageKey  string        [logger] Key for message in JSON ${KETCHUP_LOGGER_MESSAGE_KEY} (default "msg")
  --loggerTimeKey     string        [logger] Key for timestamp in JSON ${KETCHUP_LOGGER_TIME_KEY} (default "time")
  --minify                          Minify HTML ${KETCHUP_MINIFY} (default true)
  --name              string        [server] Name ${KETCHUP_NAME} (default "http")
  --okStatus          int           [http] Healthy HTTP Status code ${KETCHUP_OK_STATUS} (default 204)
  --pathPrefix        string        Root Path Prefix ${KETCHUP_PATH_PREFIX}
  --port              uint          [server] Listen port (0 to disable) ${KETCHUP_PORT} (default 1080)
  --pprofAgent        string        [pprof] URL of the Datadog Trace Agent (e.g. http://datadog.observability:8126) ${KETCHUP_PPROF_AGENT}
  --pprofPort         int           [pprof] Port of the HTTP server (0 to disable) ${KETCHUP_PPROF_PORT} (default 0)
  --publicURL         string        Public URL ${KETCHUP_PUBLIC_URL} (default "https://ketchup.vibioh.fr")
  --readTimeout       duration      [server] Read Timeout ${KETCHUP_READ_TIMEOUT} (default 5s)
  --redisAddress      string slice  [redis] Redis Address host:port (blank to disable) ${KETCHUP_REDIS_ADDRESS}, as a string slice, environment variable separated by "," (default [127.0.0.1:6379])
  --redisDatabase     int           [redis] Redis Database ${KETCHUP_REDIS_DATABASE} (default 0)
  --redisPassword     string        [redis] Redis Password, if any ${KETCHUP_REDIS_PASSWORD}
  --redisUsername     string        [redis] Redis Username, if any ${KETCHUP_REDIS_USERNAME}
  --shutdownTimeout   duration      [server] Shutdown Timeout ${KETCHUP_SHUTDOWN_TIMEOUT} (default 10s)
  --telemetryRate     string        [telemetry] OpenTelemetry sample rate, 'always', 'never' or a float value ${KETCHUP_TELEMETRY_RATE} (default "always")
  --telemetryURL      string        [telemetry] OpenTelemetry gRPC endpoint (e.g. otel-exporter:4317) ${KETCHUP_TELEMETRY_URL}
  --telemetryUint64                 [telemetry] Change OpenTelemetry Trace ID format to an unsigned int 64 ${KETCHUP_TELEMETRY_UINT64} (default true)
  --title             string        Application title ${KETCHUP_TITLE} (default "Ketchup")
  --url               string        [alcotest] URL to check ${KETCHUP_URL}
  --userAgent         string        [alcotest] User-Agent for check ${KETCHUP_USER_AGENT} (default "Alcotest")
  --writeTimeout      duration      [server] Write Timeout ${KETCHUP_WRITE_TIMEOUT} (default 10s)

```

## Contributing

Thanks for your interest in contributing! There are many ways to contribute to this project. [Get started here](CONTRIBUTING.md).

## CI

Following variables are required for CI:

|            Name            |          Purpose           |
| :------------------------: | :------------------------: |
| **SCRIPTS_NO_INTERACTIVE** | for disabling prompt in CI |
