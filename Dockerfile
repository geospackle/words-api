FROM golang:1.22.3-alpine as builder

WORKDIR /app

COPY . . 

RUN go mod download

RUN go build -o words-api ./src

FROM alpine:latest

COPY --from=builder /app/words-api /app/words-api

CMD ["/app/words-api"]


