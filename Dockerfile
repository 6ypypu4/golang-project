## Multi-stage build to produce lean images for both API and Admin binaries
ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata

# Download dependencies first to maximize layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

ENV CGO_ENABLED=0

FROM base AS build-api
RUN go build -o /out/api ./cmd/api

FROM base AS build-admin
RUN go build -o /out/admin ./cmd/admin

# Runtime image for API
FROM gcr.io/distroless/static:nonroot AS api
COPY --from=build-api /out/api /api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]

# Runtime image for Admin
FROM gcr.io/distroless/static:nonroot AS admin
COPY --from=build-admin /out/admin /admin
USER nonroot:nonroot
EXPOSE 8081
ENTRYPOINT ["/admin"]

