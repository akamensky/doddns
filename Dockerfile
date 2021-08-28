FROM golang:1.17-alpine3.14 AS builder
WORKDIR /code
COPY . /code
RUN CGO_ENABLED=0 GOOS=linux go build -a -v -o /doddns

FROM alpine:3.14
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /dodns ./
ENTRYPOINT "/dodns"
CMD [""]
