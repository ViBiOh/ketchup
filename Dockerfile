FROM scratch

ENV ZONEINFO zoneinfo.zip
EXPOSE 1080

ENV API_PORT 1080
ENV API_CSP "default-src 'self'; base-uri 'self'; script-src 'self' 'unsafe-inline' unpkg.com/swagger-ui-dist@3/; style-src 'self' 'unsafe-inline' unpkg.com/swagger-ui-dist@3/; img-src 'self' data:"
ENV API_SWAGGER_TITLE ketchup

HEALTHCHECK --retries=5 CMD [ "/ketchup", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/ketchup" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY zoneinfo.zip /
COPY release/ketchup_${TARGETOS}_${TARGETARCH} /ketchup
