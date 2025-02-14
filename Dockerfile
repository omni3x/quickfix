FROM    golang:1.19-buster as build

WORKDIR /home/app

COPY go.mod go.sum ./

# Pull dependencies, if mod/sum aren't changed then this is cached
RUN go mod download
RUN go mod verify

COPY . .

RUN go test

