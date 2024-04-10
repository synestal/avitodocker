FROM golang:1.14-alpine

RUN go version
ENV GOPATH=/

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o applicat .cmd/avito-test-trainee/main.go

RUN chmod +x applicat

CMD ["./applicat"]