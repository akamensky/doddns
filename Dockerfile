FROM golang:1.17-alpine3.14 AS builder
WORKDIR /code
COPY . /code
RUN CGO_ENABLED=0 GOOS=linux go build -a -v -o /doddns

FROM alpine:3.14
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /doddns ./
ENV DODDNS_HOSTNAME=""
ENV DODDNS_API_TOKEN_FILE=""
ENV DODDNS_CHECK_INTERVAL_MINUTES=""
ENV DODDNS_RECORD_TTL=""
ENV DODDNS_IP_SERVICES=""
ENTRYPOINT "/doddns"
CMD [""]
