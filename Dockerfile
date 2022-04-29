FROM golang:latest

RUN mkdir /app

WORKDIR /app

COPY . .

RUN go build -o main ./cmd/gophermart

CMD [ "/app/main", "-a", "gophermart:8080", "-d", "postgres://gopher:G0ph3R@postgres:5432/gophermart" ]