FROM golang:1.13-alpine as build

ARG GITHUB_TOKEN
ENV GITHUB_TOKEN $GITHUB_TOKEN

RUN apk update &&\
    apk add --no-cache git build-base

RUN git config --system url."https://$GITHUB_TOKEN:x-oauth-basic@github.com/".insteadOf "https://github.com/"

RUN adduser -D -g '' app

WORKDIR /home/app

COPY go.mod go.sum ./

# Pull dependencies, if mod/sum aren't changed then this is cached
RUN go mod download
RUN go mod verify

COPY . .

RUN go test
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"

