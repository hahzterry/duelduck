FROM node:24.15.0-alpine AS deps
WORKDIR /app
RUN apk add --no-cache libc6-compat python3 make g++ build-base libusb-dev eudev-dev linux-headers
COPY package*.json ./
RUN npm install

FROM deps AS builder
COPY . /app
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

FROM node:24.15.0-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/dist/standalone ./
COPY --from=builder /app/dist/static ./dist/static
COPY --from=builder /app/public ./public

# Rename Next.js standalone server, use our wrapper as entry point
# server.js patches http.createServer to normalize double-slash URLs
# with a single 301 before Next.js can emit a 308
RUN mv server.js next-server.js
COPY server.js ./server.js

EXPOSE 3000

CMD ["node", "server.js"]