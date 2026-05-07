FROM node:20-alpine

WORKDIR /app

RUN npm install -g serve

COPY package*.json ./

RUN npm ci --no-audit --no-fund --no-progress

COPY . .
RUN npm run build

RUN npm prune --production

EXPOSE 3000

CMD ["serve", "-s", "dist"]