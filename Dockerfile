# Stage 1: Build Angular frontend
FROM node:24-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build -- --configuration=production

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
# Download deps first (cached layer)
COPY go.mod go.sum ./
RUN go mod download
# Copy source (web/dist is excluded via .dockerignore)
COPY . .
# Overwrite with freshly-built frontend — must come after COPY . . so the
# builder output always wins over any stale local web/dist.
COPY --from=frontend-builder /app/web/dist/dart-scheduler/browser ./web/dist/dart-scheduler/browser
RUN CGO_ENABLED=0 go build -o dartscheduler ./cmd/server/

# Stage 3: Minimal runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=go-builder /app/dartscheduler .
EXPOSE 8080
VOLUME ["/data"]
ENV DATABASE_PATH=/data/dartscheduler.db
CMD ["./dartscheduler"]
