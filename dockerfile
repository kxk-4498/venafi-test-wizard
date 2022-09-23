FROM golang:1.19.1 as dev

WORKDIR /venafi

# FROM golang:1.15-alpine as build

# WORKDIR /app
# COPY ./app/* /app/
# RUN go build -o app

# FROM alpine as runtime 
# COPY --from=build /app/app /
# CMD ./app