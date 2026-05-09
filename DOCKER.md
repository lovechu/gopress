# GoPress Docker Deployment Guide

This guide explains how to run GoPress using Docker and Docker Compose.

## Prerequisites

- Docker Desktop (Windows/Mac) or Docker Engine (Linux)
- Docker Compose v1.28+ or Docker Compose plugin

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd gopress
```

### 2. Configure Environment Variables

Copy the example environment file and modify it as needed:

```bash
cp .env.example .env
```

Edit `.env` and set your configuration:

```env
# Required: Change these for production
APP_SECRET_KEY=your-super-secret-key-here
DB_PASSWORD=your-secure-database-password
REDIS_PASSWORD=your-redis-password
JWT_ACCESS_SECRET=your-jwt-access-secret
JWT_REFRESH_SECRET=your-jwt-refresh-secret

# Optional: Customize ports and settings
APP_PORT=8080
DB_PORT=5432
REDIS_PORT=6379
```

### 3. Run with Docker Compose

#### Development Mode (with hot reload)

```bash
docker-compose -f docker-compose.yml -f docker-compose.override.yml up
```

#### Production Mode

```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

#### Basic Mode (without optional services)

```bash
docker-compose up -d
```

#### With MinIO Object Storage

```bash
docker-compose --profile with-minio up -d
```

## Services

### Core Services

| Service    | Description                              | Default Port |
|------------|------------------------------------------|--------------|
| app        | GoPress application                       | 8080         |
| postgres   | PostgreSQL database                       | 5432         |
| redis      | Redis cache                               | 6379         |

### Optional Services

| Service      | Description                              | Default Port | Profile        |
|--------------|------------------------------------------|--------------|----------------|
| minio        | MinIO object storage server               | 9000         | with-minio     |
| minio-init   | MinIO bucket initialization               | -            | with-minio     |
| nginx        | Nginx reverse proxy (production)          | 80/443       | with-nginx     |

## Configuration

### Environment Variables

The application supports configuration via environment variables. The `docker-entrypoint.sh` script automatically generates `config.yaml` from environment variables if it doesn't exist.

Key environment variables:

| Variable              | Description                          | Default               |
|-----------------------|--------------------------------------|-----------------------|
| `APP_ENV`             | Application environment               | development           |
| `APP_PORT`            | Application port                     | 8080                  |
| `APP_DEBUG`           | Enable debug mode                    | true                  |
| `DB_DRIVER`           | Database driver                       | postgres              |
| `DB_HOST`             | Database host                         | postgres              |
| `DB_PORT`             | Database port                         | 5432                  |
| `DB_NAME`             | Database name                         | gopress               |
| `DB_USER`             | Database user                         | gopress               |
| `DB_PASSWORD`         | Database password                      | -                     |
| `REDIS_ADDR`          | Redis address                         | redis:6379            |
| `REDIS_PASSWORD`      | Redis password                        | -                     |
| `STORAGE_DRIVER`      | Storage driver (local/minio)         | local                 |
| `MEDIA_STORAGE`       | Media storage type                    | local                 |
| `JWT_ACCESS_SECRET`   | JWT access token secret               | -                     |
| `JWT_REFRESH_SECRET`  | JWT refresh token secret              | -                     |

### Using Custom config.yaml

If you prefer to use a custom `config.yaml`, mount it as a volume:

```yaml
# In docker-compose.yml or docker-compose.override.yml
services:
  app:
    volumes:
      - ./config.yaml:/app/config.yaml:ro
```

## Volumes

The following Docker volumes are created:

| Volume           | Description                          | Used By      |
|------------------|--------------------------------------|--------------|
| postgres_data    | PostgreSQL data                      | postgres     |
| redis_data       | Redis data                           | redis        |
| minio_data       | MinIO data                           | minio        |
| app_storage      | Application uploads                  | app          |
| app_logs         | Application logs                     | app          |

To preserve data, these volumes persist across container restarts.

## Common Commands

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f postgres
```

### Stop Services

```bash
docker-compose down
```

### Stop and Remove Volumes

```bash
docker-compose down -v
```

### Rebuild After Code Changes

```bash
docker-compose up -d --build
```

### Execute Command in Running Container

```bash
docker-compose exec app sh
docker-compose exec postgres psql -U gopress -d gopress
```

### Check Service Health

```bash
docker-compose ps
```

## Production Deployment

For production deployment:

1. **Use production compose file:**
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
   ```

2. **Set strong passwords** in `.env` file

3. **Enable Nginx reverse proxy** (optional):
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.prod.yml --profile with-nginx up -d
   ```

4. **Configure SSL/TLS** in the Nginx configuration

5. **Set up log rotation** (already configured in production profile)

## Troubleshooting

### Database Connection Issues

Check if PostgreSQL is healthy:
```bash
docker-compose ps postgres
docker-compose logs postgres
```

### Application Won't Start

Check application logs:
```bash
docker-compose logs app
```

### Permission Issues

If you see permission errors with mounted volumes:
```bash
sudo chown -R 1000:1000 storage/ logs/
```

### Port Conflicts

If ports are already in use, change them in `.env`:
```env
APP_PORT=8081
DB_PORT=5433
REDIS_PORT=6380
```

## Migration from MySQL to PostgreSQL

This Docker setup uses PostgreSQL instead of MySQL. To migrate:

1. Export data from MySQL:
   ```bash
   mysqldump -u root -p gopress > backup.sql
   ```

2. Convert and import to PostgreSQL (using tools like `pgloader`)

3. Update `config.yaml` or environment variables to use `postgres` driver

## File Structure

```
gopress/
├── Dockerfile                    # Multi-stage build
├── .dockerignore                 # Docker ignore file
├── docker-compose.yml            # Main compose file
├── docker-compose.override.yml   # Development overrides
├── docker-compose.prod.yml       # Production overrides
├── docker-entrypoint.sh          # Entrypoint script
├── .env.example                  # Example environment file
├── config.yaml                   # Application config
└── ...
```

## Support

For issues and questions:
- Check the logs: `docker-compose logs -f`
- Verify health: `docker-compose ps`
- Review configuration: `cat .env`

## License

Same as the GoPress project license.
