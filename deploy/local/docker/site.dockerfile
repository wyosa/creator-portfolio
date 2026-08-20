FROM node:22-alpine AS build
WORKDIR /src
# proxy target is baked into the nitro build (routeRules are build-time)
ARG NUXT_API_PROXY=http://api:8080
ENV NUXT_TELEMETRY_DISABLED=1 NUXT_API_PROXY=$NUXT_API_PROXY
COPY apps/site/package*.json ./
RUN npm ci || npm install
COPY apps/site/ ./
RUN npm run build

FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production NUXT_TELEMETRY_DISABLED=1
COPY --from=build /src/.output ./.output
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
