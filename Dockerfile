FROM golang:alpine as builder
WORKDIR /build

RUN apk add git
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o quotes-tgbot .

FROM alpine
WORKDIR /app
COPY --from=builder /build/quotes-tgbot /app/
ENV TELEGRAM_APITOKEN ""
ENTRYPOINT ["./quotes-tgbot"]