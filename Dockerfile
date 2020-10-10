FROM vibioh/scratch

ENV ZONEINFO zoneinfo.zip
ENV KETCHUP_PORT 1080

EXPOSE 1080

COPY templates/ /templates
COPY static/ /static

HEALTHCHECK --retries=5 CMD [ "/ketchup", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/ketchup" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY cacert.pem /etc/ssl/certs/ca-certificates.crt
COPY zoneinfo.zip /
COPY release/ketchup_${TARGETOS}_${TARGETARCH} /ketchup
