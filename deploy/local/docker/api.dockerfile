# golang:1.25-alpine
FROM golang:1.25-alpine@sha256:1ae0735f00daffa3aaf1363a5184c0d2dc55c78e3db4ec70241cdac97bf84b59 AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# alpine:3.21
FROM alpine:3.21@sha256:48b0309ca019d89d40f670aa1bc06e426dc0931948452e8491e3d65087abc07d
LABEL org.opencontainers.image.title="portfolio-api" \
      org.opencontainers.image.description="Go/Gin API for the creator portfolio (public content + admin panel)"
RUN adduser -D -u 10001 app && mkdir -p /data/media && chown -R app:app /data
USER app
COPY --from=build /out/api /bin/api
ENV PORT=8080 DATA_DIR=/data GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["/bin/api"]
