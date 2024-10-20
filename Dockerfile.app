FROM golang:1.23-alpine as builder

WORKDIR /app
COPY . .

CMD [ "go", "run", "." ]