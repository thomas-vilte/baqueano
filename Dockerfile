# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27.3

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build

WORKDIR /src

COPY go.* ./

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY prompts/ ./prompts/

ARG TARGETOS
ARG TARGETARCH
ARG BIN=agent

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /out/baqueano ./cmd/${BIN}

FROM gcr.io/distroless/static-debian12:nonroot

LABEL org.opencontainers.image.source="https://github.com/thomas-vilte/baqueano"

COPY --from=build /out/baqueano /baqueano

ENTRYPOINT ["/baqueano"]
