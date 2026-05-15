FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o builder-service ./cmd/builder/main.go


FROM alpine:latest AS certs
RUN apk --no-cache add ca-certificates tzdata


FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=certs /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=certs /etc/passwd /etc/passwd
COPY --from=certs /etc/group /etc/group

WORKDIR /app/

COPY --from=builder /app/builder-service .

USER nobody:nobody

CMD ["./builder-service"]