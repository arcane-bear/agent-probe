FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /agent-probe ./cmd/agent-probe

FROM alpine:3.19

RUN apk --no-cache add ca-certificates
COPY --from=builder /agent-probe /usr/local/bin/agent-probe

EXPOSE 8089
ENTRYPOINT ["agent-probe"]
CMD ["--config", "/etc/agent-probe/config.yaml"]
