# Stage 1: Build production React assets
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies using pnpm
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

# Copy source code and build static assets
COPY . .
RUN pnpm build

# Stage 2: Serve static files with Nginx
FROM nginx:alpine

# Copy built assets to Nginx web root
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose HTTP port
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]