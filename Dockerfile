FROM golang:alpine AS waiter

RUN apk update && apk add git

RUN go get github.com/adrian-gheorghe/wait-go

WORKDIR /go/src/github.com/adrian-gheorghe/wait-go

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/wait-go
