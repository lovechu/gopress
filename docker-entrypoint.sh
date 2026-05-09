#!/bin/sh
# Entrypoint script for GoPress Docker container
# Generates config.yaml from environment variables if not provided

set -e

# Default config file path
CONFIG_FILE="/app/config.yaml"

# If config.yaml doesn't exist, generate it from environment variables
if [ ! -f "$CONFIG_FILE" ]; then
    echo "Generating config.yaml from environment variables..."

    cat > "$CONFIG_FILE" << EOF
# GoPress CMS Configuration (Auto-generated from environment variables)

app:
  name: "GoPress"
  env: "${APP_ENV:-development}"
  port: "${APP_PORT:-8080}"
  base_url: "${APP_BASE_URL:-http://localhost:8080}"
  base_path: "."
  debug: ${APP_DEBUG:-true}
  timezone: "${TZ:-Asia/Shanghai}"
  secret_key: "${APP_SECRET_KEY:-change-me-to-a-random-secret-key}"

database:
  driver: "${DB_DRIVER:-postgres}"
  host: "${DB_HOST:-postgres}"
  port: ${DB_PORT:-5432}
  name: "${DB_NAME:-gopress}"
  user: "${DB_USER:-gopress}"
  password: "${DB_PASSWORD:-gopress_secure_password}"
  charset: "utf8"
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 3600
  log_level: "warn"
  sslmode: "${DB_SSLMODE:-disable}"

redis:
  addr: "${REDIS_ADDR:-redis:6379}"
  password: "${REDIS_PASSWORD:-}"
  db: ${REDIS_DB:-0}
  pool_size: 20
  prefix: "gopress:"

jwt:
  access_secret: "${JWT_ACCESS_SECRET:-change-access-secret-key}"
  refresh_secret: "${JWT_REFRESH_SECRET:-change-refresh-secret-key}"
  access_expire: 3600
  refresh_expire: 604800

storage:
  driver: "${STORAGE_DRIVER:-local}"
  local:
    root: "./storage/uploads"
    base_url: "/uploads"

upload:
  max_size: 20971520
  allowed_types:
    - "image/jpeg"
    - "image/png"
    - "image/gif"
    - "image/webp"
    - "image/avif"
    - "application/pdf"
  image_quality: 85
  generate_webp: true

media:
  storage: "${MEDIA_STORAGE:-local}"
  max_file_size: 52428800
  allowed_types:
    - "image/jpeg"
    - "image/png"
    - "image/gif"
    - "image/webp"
    - "application/pdf"
    - "application/msword"
    - "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
  local:
    base_path: "./storage/uploads"
    base_url: "/uploads"
EOF

    # Add MinIO configuration if using MinIO
    if [ "$MEDIA_STORAGE" = "minio" ] || [ "$STORAGE_DRIVER" = "minio" ]; then
        cat >> "$CONFIG_FILE" << EOF
  minio:
    endpoint: "${MINIO_ENDPOINT:-minio:9000}"
    access_key: "${MINIO_ACCESS_KEY:-minioadmin}"
    secret_key: "${MINIO_SECRET_KEY:-minioadmin}"
    bucket: "${MINIO_BUCKET:-gopress}"
    use_ssl: ${MINIO_USE_SSL:-false}
    region: "${MINIO_REGION:-us-east-1}"
    base_url: "http://${MINIO_ENDPOINT:-minio:9000}/${MINIO_BUCKET:-gopress}"
EOF
    fi

    cat >> "$CONFIG_FILE" << EOF

cache:
  driver: "redis"
  default_ttl: 300

log:
  level: "${LOG_LEVEL:-info}"
  format: "json"
  output: "stdout"

cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "PATCH"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Authorization"
    - "Content-Type"
    - "X-Request-ID"
  max_age: 86400

rate_limit:
  enabled: true
  requests: 100
  window: 60
EOF

    echo "config.yaml generated successfully."
fi

# Execute the main command
exec "$@"
