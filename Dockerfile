FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum /app/
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o licenser ./cmd/licenser/main.go

FROM gcr.io/distroless/static-debian12:latest
USER 65532:65532

WORKDIR /app
COPY --from=builder /app/licenser .

CMD [ "/app/licenser" ]
