# node:22-alpine
FROM node:22-alpine@sha256:c610fcdfb1d5b4740dd70c284ed3cb16bb857e0f7166196e36a5501df7a3aa32 AS build
WORKDIR /src
# proxy target is baked into the nitro build (routeRules are build-time)
ARG NUXT_API_PROXY=http://api:8080
ENV NUXT_TELEMETRY_DISABLED=1 NUXT_API_PROXY=$NUXT_API_PROXY
COPY apps/site/package*.json ./
RUN npm ci
COPY apps/site/ ./
RUN npm run build

# node:22-alpine
FROM node:22-alpine@sha256:c610fcdfb1d5b4740dd70c284ed3cb16bb857e0f7166196e36a5501df7a3aa32
LABEL org.opencontainers.image.title="portfolio-site" \
      org.opencontainers.image.description="Nuxt 3 frontend for the creator portfolio (proxies /api and /media to the api)"
WORKDIR /app
ENV NODE_ENV=production NUXT_TELEMETRY_DISABLED=1
COPY --from=build --chown=node:node /src/.output ./.output
USER node
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
