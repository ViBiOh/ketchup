FROM golang:alpine as build
RUN apk --no-cache add tzdata ca-certificates

FROM vibioh/scratch

ENV KETCHUP_PORT 1080
EXPOSE 1080

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

COPY templates/ /templates
COPY static/ /static

HEALTHCHECK --retries=5 CMD [ "/ketchup", "-url", "http://localhost:1080/health" ]
ENTRYPOINT [ "/ketchup" ]

ARG VERSION
ENV VERSION=${VERSION}

ARG TARGETOS
ARG TARGETARCH

COPY release/ketchup_${TARGETOS}_${TARGETARCH} /ketchup
