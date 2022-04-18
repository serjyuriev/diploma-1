FROM golang:latest

RUN mkdir /app

WORKDIR /app

COPY . .

RUN go build -o main ./cmd/gophermart

CMD [ "/app/main" ]