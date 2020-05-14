FROM golang:1.14 AS builder
WORKDIR /go/src/github.com/polyse/database

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -installsuffix cgo -o app github.com/polyse/database/cmd/database

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=0 /go/src/github.com/polyse/database .
CMD mkdir /var/data
ENV DB_FILE /var/data
ENV LOG_FMT json
ENTRYPOINT ["/app/app"]