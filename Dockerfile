# Global build arguments for base image versions
ARG VERSION=latest
ARG NODE_VERSION=22
ARG TEMPL_VERSION=v0.3.943
ARG GO_VERSION=1.25
ARG ALPINE_VERSION=3.21
ARG TARGETPLATFORM

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
# Build Tailwindcss
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
FROM node:${NODE_VERSION}-alpine AS ui
ENV NODE_ENV=production
WORKDIR /app
COPY package*.json ./
RUN npm ci --omit=dev
COPY ui/ ui/
RUN npx @tailwindcss/cli -i ./ui/static/css/input.css -o ./ui/static/css/output.css --minify

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
# Build templ file
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
FROM ghcr.io/a-h/templ:${TEMPL_VERSION} AS templ
WORKDIR /app
COPY --chown=65532:65532 . .
RUN ["templ", "generate"]

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
# Build Go project
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
FROM golang:${GO_VERSION}-alpine AS build

# Re-declare ARG for use in this stage

ARG VERSION
ARG TARGETOS
ARG TARGETARCH
ARG CGO_ENABLED=1
ARG GOFLAGS="-buildvcs=false"

ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV GOFLAGS=${GOFLAGS}

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY --from=templ /app .
RUN go build -ldflags="-s -w -X main.version=${VERSION}" -o main ./cmd/web && \
    go build -ldflags="-s -w -X main.version=${VERSION}" -o gocost ./cmd/cli

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
# Final stage
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
FROM alpine:${ALPINE_VERSION}
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite \
    tzdata \
    wget && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binaries and configuration
COPY --from=build /app/main ./main
COPY --from=build /app/gocost ./gocost
COPY --from=ui /app/ui/static ./ui/static
COPY --from=build /app/migrations ./migrations
COPY docker-entrypoint.sh /usr/local/bin/

# Set up data and uploads directories with proper permissions
RUN mkdir -p ./data ./uploads && \
    chown -R appuser:appgroup /app && \
    chmod +x /usr/local/bin/docker-entrypoint.sh

# Switch to non-root user
USER appuser

# Set environment variables
ENV DB_PATH=/app/data/data.sqlite

EXPOSE 4000

# Health check to monitor application status
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4000/ || exit 1

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
