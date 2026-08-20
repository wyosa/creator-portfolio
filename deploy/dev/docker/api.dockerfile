FROM golang:1.25-alpine

WORKDIR /app
RUN go install github.com/air-verse/air@v1.67.4

CMD ["air", "-c", ".air.toml"]
