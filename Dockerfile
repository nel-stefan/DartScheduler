# Stage 1: Build Angular frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build -- --configuration=production

# Stage 2: Build Go binary
FROM golang:1.26-alpine AS go-builder
WORKDIR /app
# Copy Angular output into the embed path
COPY --from=frontend-builder /app/web/dist/dart-scheduler/browser ./web/dist/dart-scheduler/browser
# Download deps
COPY go.mod go.sum ./
RUN go mod download
# Copy source
COPY . .
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
