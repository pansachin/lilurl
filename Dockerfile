FROM golang:1.23-alpine as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=off go build -o lilurl .

CMD [ "./lilurl" ]
