# node:22-alpine
FROM node:22-alpine@sha256:c610fcdfb1d5b4740dd70c284ed3cb16bb857e0f7166196e36a5501df7a3aa32

WORKDIR /app
COPY apps/site/package*.json ./
RUN npm ci
COPY apps/site/ ./

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
