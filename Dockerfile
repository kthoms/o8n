# -----------------------------------------------------------------------------
# The base image for building the o8n binary
FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine3.21 AS build

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH

WORKDIR /o8n
COPY go.mod go.sum main.go Makefile ./
RUN apk --no-cache add --update make libx11-dev git gcc libc-dev curl \
  && make build

# -----------------------------------------------------------------------------
# Build the final Docker image
FROM --platform=$BUILDPLATFORM alpine:3.23.3
ARG KUBECTL_VERSION="v1.32.2"

COPY --from=build /o8n/execs/o8n /bin/o8n
RUN apk --no-cache add --update ca-certificates \
  && apk --no-cache add --update -t deps curl vim \
  && TARGET_ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/) \
  && apk del --purge deps

ENTRYPOINT [ "/bin/o8n" ]