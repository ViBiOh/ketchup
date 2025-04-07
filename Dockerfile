FROM rg.fr-par.scw.cloud/vibioh/scratch

ENV KETCHUP_PORT 1080
EXPOSE 1080

ENV ZONEINFO /zoneinfo.zip
COPY zoneinfo.zip /zoneinfo.zip
COPY ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

HEALTHCHECK --retries=5 CMD [ "/ketchup", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/ketchup" ]

ARG VERSION
ENV VERSION ${VERSION}

ARG GIT_SHA
ENV GIT_SHA ${GIT_SHA}

ARG TARGETOS
ARG TARGETARCH

COPY release/ketchup_${TARGETOS}_${TARGETARCH} /ketchup
COPY release/notifier_${TARGETOS}_${TARGETARCH} /notifier
