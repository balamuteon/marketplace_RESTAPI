FROM golang:1.24.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/app/main.go


FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache curl ca-certificates

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

COPY --from=builder /server /app/server

COPY ./migrations /app/migrations

COPY config.yaml .


EXPOSE 8080

CMD [ "/app/server" ]