# golang:1.25-alpine
FROM golang:1.25-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59

WORKDIR /app
RUN go install github.com/air-verse/air@v1.67.4

CMD ["air", "-c", ".air.toml"]
