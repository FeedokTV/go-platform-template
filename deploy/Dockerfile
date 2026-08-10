FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Why ARGs after mod download: version changes must not bust the dependency cache
ARG VERSION=dev
ARG COMMIT=dev
ARG DATE=dev

RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w \
      -X go-platform-template/internal/platform/buildinfo.Version=${VERSION} \
      -X go-platform-template/internal/platform/buildinfo.Commit=${COMMIT} \
      -X go-platform-template/internal/platform/buildinfo.Date=${DATE}" \
    -o /server ./cmd

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /server /server
EXPOSE 8080

ENTRYPOINT ["/server"]