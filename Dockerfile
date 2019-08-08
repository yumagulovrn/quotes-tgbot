FROM golang:alpine as builder
WORKDIR /build

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o quotes-tgbot .

FROM alpine
WORKDIR /app
COPY --from=builder /build/quotes-tgbot /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV TELEGRAM_APITOKEN ""
ENTRYPOINT ["./quotes-tgbot"]