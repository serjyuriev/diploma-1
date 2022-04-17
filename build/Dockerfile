FROM golang:latest

RUN go install github.com/canthefason/go-watcher/cmd/watcher@latest

WORKDIR /go/src/github.com/serjyuriev/diploma-1

COPY .. .

WORKDIR /go/src/github.com/serjyuriev/diploma-1/cmd/gophermart

ENTRYPOINT /go/bin/watcher