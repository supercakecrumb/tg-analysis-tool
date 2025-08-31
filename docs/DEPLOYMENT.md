# Deployment Guide

## Overview

This guide covers deploying tg-analysis-tool to a home server using Docker Compose. The application is designed to be self-hosted with minimal external dependencies.

## Prerequisites

- Linux server (Ubuntu 22.04+ or similar)
- Docker Engine 24.0+
- Docker Compose v2.20+
- 4GB+ RAM (8GB recommended)
- 50GB+ storage (depends on data volume)
- Domain name (optional, for HTTPS)
- Telegram API credentials

## Quick Start

### 1. Server Preparation

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sudo sh

# Add user to docker group
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin

# Verify installation
docker --version
docker compose version
```

### 2. Clone Repository

```bash
git clone https://github.com/supercakecrumb/tg-analysis-tool.git
cd tg-analysis-tool
```

### 3. Configure Environment

```bash
# Copy example environment
cp .env.example .env

# Generate secure keys
echo "SESSION_ENC_KEY=$(openssl rand -base64 32)" >> .env
echo "COOKIE_SECRET=$(openssl rand -hex 16)" >> .env

# Edit configuration
nano .env
```

Required configuration:
```env
# Telegram API (from https://my.telegram.org)
TELEGRAM_API_ID=your_api_id
TELEGRAM_API_HASH=your_api_hash

# Application
BASE_URL=https://your-domain.com  # or http://localhost:3000
TIMEZONE=Europe/Moscow

# Database passwords (change these!)
POSTGRES_PASSWORD=secure_password_here
CLICKHOUSE_PASSWORD=another_secure_password
```

### 4. Deploy Application

```bash
# Pull latest images
docker compose pull

# Start services
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f
```

### 5. Initialize Database

```bash
# Run migrations
docker compose exec api /app/migrate up

# Verify database
docker compose exec postgres psql -U tg_user -d tg_analysis -c "\dt"
```

## Production Configuration

### Docker Compose Configuration

`docker-compose.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: tg_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: tg_analysis
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U tg_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  clickhouse:
    image: clickhouse/clickhouse-server:23.8-alpine
    restart: unless-stopped
    environment:
      CLICKHOUSE_DB: tg_analytics
      CLICKHOUSE_USER: tg_user
      CLICKHOUSE_PASSWORD: ${CLICKHOUSE_PASSWORD}
    volumes:
      - clickhouse_data:/var/lib/clickhouse
      - ./migrations/clickhouse:/docker-entrypoint-initdb.d
    ports:
      - "127.0.0.1:8123:8123"  # HTTP interface (local only)
    ulimits:
      nofile:
        soft: 262144
        hard: 262144

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data

  api:
    image: ghcr.io/yourusername/tg-analysis-tool:latest
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      clickhouse:
        condition: service_started
      redis:
        condition: service_started
    environment:
      DATABASE_URL: postgres://tg_user:${POSTGRES_PASSWORD}@postgres:5432/tg_analysis?sslmode=disable
      CLICKHOUSE_URL: clickhouse://tg_user:${CLICKHOUSE_PASSWORD}@clickhouse:9000/tg_analytics
      REDIS_URL: redis://redis:6379
      SESSION_ENC_KEY: ${SESSION_ENC_KEY}
      COOKIE_SECRET: ${COOKIE_SECRET}
      TELEGRAM_API_ID: ${TELEGRAM_API_ID}
      TELEGRAM_API_HASH: ${TELEGRAM_API_HASH}
      BASE_URL: ${BASE_URL}
      TIMEZONE: ${TIMEZONE}
      LOG_LEVEL: info
    ports:
      - "127.0.0.1:8080:8080"
    volumes:
      - ./data:/app/data  # For file uploads/exports

  web:
    image: ghcr.io/yourusername/tg-analysis-tool-web:latest
    restart: unless-stopped
    depends_on:
      - api
    environment:
      NEXT_PUBLIC_API_URL: ${BASE_URL}/api
    ports:
      - "127.0.0.1:3000:3000"

volumes:
  postgres_data:
  clickhouse_data:
  redis_data:
```

### Nginx Reverse Proxy

Install and configure Nginx:

```bash
sudo apt install nginx certbot python3-certbot-nginx
```

Nginx configuration (`/etc/nginx/sites-available/tg-analysis`):
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;
    
    # SSL configuration (managed by Certbot)
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self' https:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';" always;
    
    # Frontend
    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # API
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support for SSE
        proxy_set_header Connection '';
        proxy_buffering off;
        proxy_cache off;
    }
    
    # Metrics (protected)
    location /metrics {
        proxy_pass http://localhost:8080;
        allow 127.0.0.1;
        deny all;
    }
}
```

Enable site and get SSL certificate:
```bash
sudo ln -s /etc/nginx/sites-available/tg-analysis /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
sudo certbot --nginx -d your-domain.com
```

## Backup and Recovery

### Automated Backups

Create backup script (`/opt/tg-analysis/backup.sh`):
```bash
#!/bin/bash
set -e

BACKUP_DIR="/backup/tg-analysis"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup PostgreSQL
docker compose exec -T postgres pg_dump -U tg_user tg_analysis | \
    gzip > "$BACKUP_DIR/postgres_$DATE.sql.gz"

# Backup ClickHouse
docker compose exec -T clickhouse clickhouse-client \
    --query "SELECT * FROM tg_analytics.messages FORMAT Native" | \
    gzip > "$BACKUP_DIR/clickhouse_messages_$DATE.native.gz"

# Backup environment
cp .env "$BACKUP_DIR/env_$DATE"

# Clean old backups
find "$BACKUP_DIR" -type f -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $DATE"
```

Add to crontab:
```bash
# Daily backup at 3 AM
0 3 * * * /opt/tg-analysis/backup.sh >> /var/log/tg-backup.log 2>&1
```

### Recovery Procedure

1. **Stop services**:
```bash
docker compose down
```

2. **Restore PostgreSQL**:
```bash
gunzip < /backup/tg-analysis/postgres_20240101_030000.sql.gz | \
    docker compose exec -T postgres psql -U tg_user tg_analysis
```

3. **Restore ClickHouse**:
```bash
gunzip < /backup/tg-analysis/clickhouse_messages_20240101_030000.native.gz | \
    docker compose exec -T clickhouse clickhouse-client \
    --query "INSERT INTO tg_analytics.messages FORMAT Native"
```

4. **Start services**:
```bash
docker compose up -d
```

## Monitoring

### Health Checks

Create monitoring script (`/opt/tg-analysis/healthcheck.sh`):
```bash
#!/bin/bash

# Check API health
API_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ "$API_STATUS" != "200" ]; then
    echo "API unhealthy: $API_STATUS"
    # Send alert
fi

# Check database connections
docker compose exec api /app/healthcheck

# Check disk space
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 80 ]; then
    echo "Disk usage critical: $DISK_USAGE%"
    # Send alert
fi
```

### Prometheus Metrics

Configure Prometheus (`prometheus.yml`):
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'tg-analysis'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

### Logging

Configure log rotation (`/etc/logrotate.d/tg-analysis`):
```
/var/log/tg-analysis/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create 0640 root root
}
```

## Security Hardening

### Firewall Configuration

```bash
# Allow SSH (change port as needed)
sudo ufw allow 22/tcp

# Allow HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable
```

### Fail2ban Configuration

```ini
# /etc/fail2ban/jail.local
[tg-analysis]
enabled = true
port = http,https
filter = tg-analysis
logpath = /var/log/nginx/access.log
maxretry = 5
bantime = 3600
```

### System Updates

Enable automatic security updates:
```bash
sudo apt install unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades
```

## Troubleshooting

### Common Issues

1. **Container won't start**:
```bash
# Check logs
docker compose logs api

# Verify environment
docker compose config

# Check resources
docker system df
```

2. **Database connection errors**:
```bash
# Test PostgreSQL
docker compose exec postgres pg_isready

# Test ClickHouse
curl http://localhost:8123/ping
```

3. **High memory usage**:
```bash
# Check container stats
docker stats

# Adjust ClickHouse memory
# In docker-compose.yml, add:
environment:
  CLICKHOUSE_MAX_MEMORY_USAGE: 2000000000  # 2GB
```

### Performance Tuning

1. **PostgreSQL tuning**:
```sql
-- Adjust shared_buffers
ALTER SYSTEM SET shared_buffers = '256MB';

-- Increase work_mem for complex queries
ALTER SYSTEM SET work_mem = '16MB';
```

2. **ClickHouse tuning**:
```xml
<!-- /etc/clickhouse-server/users.d/performance.xml -->
<profiles>
    <default>
        <max_memory_usage>4000000000</max_memory_usage>
        <max_threads>4</max_threads>
    </default>
</profiles>
```

## Maintenance

### Regular Tasks

Weekly:
- Review logs for errors
- Check disk usage
- Verify backup integrity

Monthly:
- Update Docker images
- Review security updates
- Analyze performance metrics

Quarterly:
- Test disaster recovery
- Review and rotate secrets
- Audit user access

### Upgrade Procedure

1. **Backup data**:
```bash
./backup.sh
```

2. **Pull new images**:
```bash
docker compose pull
```

3. **Run migrations**:
```bash
docker compose run --rm api /app/migrate up
```

4. **Restart services**:
```bash
docker compose up -d
```

5. **Verify functionality**:
```bash
curl http://localhost:8080/health
```

## Support

For issues and questions:
1. Check logs: `docker compose logs`
2. Review documentation
3. Open GitHub issue with:
   - Error messages
   - Docker/system versions
   - Steps to reproduce