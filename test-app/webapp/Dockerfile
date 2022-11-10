FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod localhost.crt localhost.key ./app/

COPY *.go ./

RUN go build -o /webapp
RUN apk add --no-cache bash

EXPOSE 9000

CMD [ "/webapp" ]