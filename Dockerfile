# Builder stage
FROM golang:1.19 AS builder
WORKDIR /workspace

ARG GOPROXY=https://proxy.golang.org
ENV GOPROXY=$GOPROXY

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download

COPY ./ ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} \
    go build -trimpath -ldflags "${ldflags} -extldflags '-static'" \
    -o openai-bot ${package}

# Production stage
FROM gcr.io/distroless/base-debian11
WORKDIR /
COPY --from=builder /workspace/openai-bot /openai-bot
USER 65530
CMD ["/openai-bot"]
