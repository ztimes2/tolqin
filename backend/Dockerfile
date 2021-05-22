FROM golang:1.16.4-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY internal/ internal
COPY vendor/ vendor
COPY cmd/tolqin-api/ cmd/tolqin-api

RUN go build -o app -mod vendor cmd/tolqin-api/main.go

FROM alpine:3.13.5

COPY --from=builder /app/app .

ENV SERVER_PORT 80

EXPOSE 80

CMD ./app