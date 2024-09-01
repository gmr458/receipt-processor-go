##
## Build
##
FROM golang:1.22.5-alpine3.20 AS builder

WORKDIR /usr/src/app/

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -ldflags '-s -w' -o webservice ./cmd/webservice

##
## Deploy
##
FROM alpine:3.20.2

WORKDIR /app/

COPY --from=builder /usr/src/app/webservice /app/webservice

EXPOSE 4000 4001

CMD ["/app/webservice"]
