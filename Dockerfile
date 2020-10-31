FROM vibioh/scratch

ENV KETCHUP_PORT 1080
EXPOSE 1080

ENV ZONEINFO /zoneinfo.zip
COPY zoneinfo.zip /zoneinfo.zip

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt

COPY templates/ /templates
COPY static/ /static

HEALTHCHECK --retries=5 CMD [ "/ketchup", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/ketchup" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY release/ketchup_${TARGETOS}_${TARGETARCH} /ketchup
