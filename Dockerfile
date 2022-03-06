# syntax=docker/dockerfile:1

FROM golang:1.17-alpine as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

# Install go modules
RUN go mod download

COPY . .

RUN go build -o /backup-monitor

FROM alpine

COPY --from=builder /backup-monitor /

EXPOSE 8080

CMD [ "/backup-monitor" ]