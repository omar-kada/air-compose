# ----------- Builder Stage -----------
FROM golang:1.25.1-bookworm AS builder

RUN mkdir /air-compose
WORKDIR /air-compose

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend .
RUN go build -o air-compose /air-compose/cmd/air-compose/main.go

# ----------- Frontend Builder Stage -----------
FROM node:22-alpine AS frontend-builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend .
RUN npm run build

# ----------- Production Stage -----------
FROM debian:bookworm-slim AS production

RUN apt update && apt install --yes --no-install-recommends \
    curl \
    ca-certificates \
    gnupg \
    unzip \
    dumb-init \
    && install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && chmod a+r /etc/apt/keyrings/docker.gpg \
    && echo \
         "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
         "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
         tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt update \
    && apt --yes --no-install-recommends install \
         docker-ce-cli \
         docker-compose-plugin \
    && rm -rf /var/lib/apt/lists/*

# Move to working directory /build

ARG UID=1000
ARG GID=1000

RUN mkdir /app
RUN mkdir /data

WORKDIR /app

COPY --from=builder /air-compose/air-compose /app/
COPY --from=frontend-builder /app/dist /app/frontend/dist

RUN chmod -R 744 /app

ENV AIR_COMPOSE_WORKING_DIR="/data"
EXPOSE 5005

# Start the application
CMD ["sh", "-c", "/app/air-compose run"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:5005/ || exit 1