FROM golang:alpine AS waiter

ENV SRC_DIR $GOPATH/src/github.com/hallelujah-shih/wait-go

RUN apk update && apk add git

WORKDIR $SRC_DIR

COPY . $SRC_DIR

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/wait-go

CMD ["/go/bin/wait-go", "-h"]
