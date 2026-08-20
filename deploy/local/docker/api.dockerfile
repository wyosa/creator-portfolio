FROM golang:1.25-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.21
RUN adduser -D -u 10001 app && mkdir -p /data/media && chown -R app:app /data
USER app
COPY --from=build /out/api /bin/api
ENV PORT=8080 DATA_DIR=/data GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
