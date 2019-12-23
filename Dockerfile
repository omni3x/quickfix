FROM golang:1.13-alpine as build

ARG GITHUB_TOKEN
ENV GITHUB_TOKEN $GITHUB_TOKEN

RUN echo "APK UPDATE"
RUN apk update &&\
    apk add --no-cache git build-base
RUN echo "FINISHED APK UPDATE"

RUN echo "GIT CONFIG"
RUN git config --system url."https://$GITHUB_TOKEN:x-oauth-basic@github.com/".insteadOf "https://github.com/"
RUN echo "FINISHED GIT CONFIG"

# Pull dependencies, if mod/sum aren't changed then this is cached
RUN echo "GO MOD DOWNLOAD"
RUN go mod download
RUN echo "GO MOD VERIFY"
RUN go mod verify

RUN echo "GO TEST"
RUN go test

