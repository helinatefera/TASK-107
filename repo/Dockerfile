FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /chargeops ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /chargeops /chargeops
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/chargeops"]
