# Docker Deployment Guide

This guide explains how to run Buck It Up using Docker and Docker Compose.

## Prerequisites

- Docker Engine 20.10 or later
- Docker Compose 1.29 or later (or Docker Compose V2)

## Quick Start

### Using Docker Compose (Recommended)

1. **Build and start the container:**
   ```bash
   docker-compose up -d
   ```

2. **View logs:**
   ```bash
   docker-compose logs -f
   ```

3. **Stop the container:**
   ```bash
   docker-compose down
   ```

4. **Stop and remove volumes (deletes all data):**
   ```bash
   docker-compose down -v
   ```

### Using Docker Directly

1. **Build the image:**
   ```bash
   docker build -t buck-it-up:latest .
   ```

2. **Run the container:**
   ```bash
   docker run -d \
     --name buck-it-up \
     -p 8080:8080 \
     -v buck-data:/app/data \
     buck-it-up:latest
   ```

3. **View logs:**
   ```bash
   docker logs -f buck-it-up
   ```

4. **Stop the container:**
   ```bash
   docker stop buck-it-up
   docker rm buck-it-up
   ```

## Configuration

### Environment Variables

You can customize the application using environment variables in `docker-compose.yml`:

- `PORT`: HTTP server port (default: 8080)
- `BUCK_DB_PATH`: Path to SQLite database file (default: /app/data/data.db)
- `BUCK_DATA_PATH`: Root directory for bucket storage (default: /app/data)

Example:
```yaml
environment:
  - PORT=8080
  - BUCK_DB_PATH=/app/data/data.db
  - BUCK_DATA_PATH=/app/data
```

### Ports

By default, the application is exposed on port 8080. To change the host port:

```yaml
ports:
  - "3000:8080"  # Access on localhost:3000
```

### Data Persistence

Data is persisted using Docker volumes. The `buck-data` volume stores:
- SQLite database (`data.db`)
- Uploaded bucket objects

To backup your data:
```bash
docker run --rm -v buck-data:/data -v $(pwd):/backup alpine tar czf /backup/buck-data-backup.tar.gz -C /data .
```

To restore from backup:
```bash
docker run --rm -v buck-data:/data -v $(pwd):/backup alpine sh -c "cd /data && tar xzf /backup/buck-data-backup.tar.gz"
```

## Accessing the Application

Once running, access the web UI at:
- http://localhost:8080

## Health Checks

The container includes a health check that runs every 30 seconds. Check status:
```bash
docker ps
```

Look for "healthy" in the STATUS column.

## Troubleshooting

### View container logs:
```bash
docker-compose logs -f buck-it-up
```

### Check if container is running:
```bash
docker-compose ps
```

### Restart the container:
```bash
docker-compose restart
```

### Rebuild after code changes:
```bash
docker-compose up -d --build
```

### Access container shell:
```bash
docker-compose exec buck-it-up sh
```

## Production Considerations

For production deployments:

1. **Use specific image tags** instead of `latest`
2. **Set up reverse proxy** (nginx, Traefik) for HTTPS
3. **Configure resource limits:**
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1'
         memory: 512M
       reservations:
         cpus: '0.5'
         memory: 256M
   ```
4. **Set up regular backups** of the data volume
5. **Use Docker secrets** for sensitive configuration
6. **Monitor logs** with a logging driver (e.g., json-file, syslog)

## Multi-Architecture Builds

To build for multiple architectures (e.g., ARM64 for Raspberry Pi):

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t buck-it-up:latest .
```

