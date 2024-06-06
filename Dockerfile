FROM golang:1.22-alpine

RUN mkdir /usr/application
WORKDIR /usr/application

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

CMD ["/docker-gs-ping"]

