FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

RUN apk add --no-cache build-base && \
  CGO_ENABLED=1 go build -o lilurl .

CMD [ "./lilurl" ]
