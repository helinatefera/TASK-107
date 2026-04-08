FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
RUN go install github.com/tsenart/vegeta@latest

FROM alpine:3.20
RUN apk add --no-cache bash curl jq
COPY --from=builder /root/go/bin/vegeta /usr/local/bin/vegeta
COPY run_load_test.sh /run_load_test.sh
RUN chmod +x /run_load_test.sh

ENTRYPOINT ["/run_load_test.sh"]
