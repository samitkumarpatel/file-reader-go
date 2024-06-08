FROM golang:1.21-alpine

RUN mkdir /usr/application
WORKDIR /usr/application

COPY go.mod go.sum ./
RUN go mod download

COPY file-processor.go ./
RUN go build -o file-processor

CMD ["./file-processor"]
