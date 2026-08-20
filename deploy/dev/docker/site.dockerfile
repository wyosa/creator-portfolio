FROM node:22-alpine

WORKDIR /app
COPY apps/site/package*.json ./
RUN npm install
COPY apps/site/ ./

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
