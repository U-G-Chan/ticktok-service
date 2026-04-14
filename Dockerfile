FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Specify the service to build via build argument
ARG SERVICE_NAME

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/${SERVICE_NAME}/main.go

FROM alpine:latest

WORKDIR /app

# Install necessary system dependencies (ffmpeg is required for media-worker, tzdata for timezones)
# We also run apk upgrade to patch any existing vulnerabilities in the alpine base image.
RUN apk update && apk upgrade && apk add --no-cache ffmpeg tzdata ca-certificates

# Copy the built binary
COPY --from=builder /app/main .

# Ensure config directory exists and copy the template as the default config.
# Real values will be injected via Kubernetes ConfigMap and Secret as environment variables.
RUN mkdir -p config
COPY --from=builder /app/config/config.yaml.template ./config/config.yaml

CMD ["./main"]
