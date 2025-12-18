# Deployment Guide

Complete guide for deploying PriceMap-Go in various environments.

## 🐳 Docker Deployment (Recommended)

### Production Docker Compose

```bash
# Start all services in production mode
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# View logs
docker-compose logs -f

# Scale services
docker-compose up -d --scale scraper=3
```

### Single Container

```bash
# Build optimized image
docker build -t pricemap-go:2.0 .

# Run server
docker run -d \
  --name pricemap-server \
  -p 3000:3000 \
  -e DB_HOST=your-db-host \
  -e USE_TOR=true \
  pricemap-go:2.0

# Run scraper (one-time)
docker run --rm \
  --name pricemap-scraper \
  -e DB_HOST=your-db-host \
  pricemap-go:2.0 /scraper
```

---

## ☸️ Kubernetes Deployment

Full Kubernetes manifests in `k8s/` directory.

### Quick Deploy

```bash
# Create namespace and deploy
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get all -n pricemap

# Get service URL
kubectl get svc pricemap-server-service -n pricemap
```

### Features

- ✅ Auto-scaling (2-10 replicas)
- ✅ Health checks (liveness/readiness)
- ✅ Resource limits
- ✅ Persistent storage for PostgreSQL
- ✅ CronJob for periodic scraping
- ✅ Multiple Tor instances

See [k8s/README.md](k8s/README.md) for details.

---

## 🌐 Cloud Platforms

### Google Cloud (GKE)

```bash
# Create GKE cluster
gcloud container clusters create pricemap-cluster \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n1-standard-2 \
  --enable-autoscaling \
  --min-nodes 2 \
  --max-nodes 10

# Get credentials
gcloud container clusters get-credentials pricemap-cluster --zone us-central1-a

# Deploy
kubectl apply -f k8s/deployment.yaml

# Get LoadBalancer IP
kubectl get svc pricemap-server-service -n pricemap -w
```

### AWS (EKS)

```bash
# Create EKS cluster
eksctl create cluster \
  --name pricemap-cluster \
  --region us-east-1 \
  --nodegroup-name standard-workers \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 2 \
  --nodes-max 10 \
  --managed

# Deploy
kubectl apply -f k8s/deployment.yaml

# Setup LoadBalancer
kubectl get svc pricemap-server-service -n pricemap
```

### Azure (AKS)

```bash
# Create AKS cluster
az aks create \
  --resource-group pricemap-rg \
  --name pricemap-cluster \
  --node-count 3 \
  --enable-cluster-autoscaler \
  --min-count 2 \
  --max-count 10 \
  --node-vm-size Standard_D2_v2 \
  --generate-ssh-keys

# Get credentials
az aks get-credentials --resource-group pricemap-rg --name pricemap-cluster

# Deploy
kubectl apply -f k8s/deployment.yaml
```

---

## 🖥️ VPS / Bare Metal

### Using Docker Compose

```bash
# Install Docker & Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Clone repository
git clone https://github.com/VadimToptunov/pricemap-go.git
cd pricemap-go

# Configure
cp env.example .env
nano .env  # Edit configuration

# Start services
docker-compose up -d

# Setup systemd service (optional)
sudo nano /etc/systemd/system/pricemap.service
```

### Systemd Service

```ini
[Unit]
Description=PriceMap-Go Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/pricemap-go
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable pricemap
sudo systemctl start pricemap
sudo systemctl status pricemap
```

### Using Binary

```bash
# Build
make build

# Run server
./bin/server &

# Run scraper (cron)
0 */6 * * * /opt/pricemap-go/bin/scraper >> /var/log/pricemap-scraper.log 2>&1
```

---

## 🔒 Security Hardening

### 1. Database

```bash
# Change default password
DB_PASSWORD=$(openssl rand -base64 32)

# Use SSL
DB_SSLMODE=require

# Restrict access
# In pg_hba.conf:
host    pricemap    pricemap    10.0.0.0/8    scram-sha-256
```

### 2. API

```bash
# Enable HTTPS (use reverse proxy)
# Nginx example:
server {
    listen 443 ssl http2;
    server_name api.pricemap.com;
    
    ssl_certificate /etc/letsencrypt/live/api.pricemap.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.pricemap.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Add rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req zone=api burst=20 nodelay;
```

### 3. Firewall

```bash
# UFW (Ubuntu)
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 443/tcp     # HTTPS
sudo ufw deny 3000/tcp     # Block direct API access
sudo ufw enable

# iptables
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -p tcp --dport 3000 -j DROP
```

### 4. Docker Security

```bash
# Run as non-root
USER 1000:1000

# Read-only filesystem
docker run --read-only pricemap-go:2.0

# Drop capabilities
docker run --cap-drop=ALL pricemap-go:2.0

# Security scanning
docker scan pricemap-go:2.0
```

---

## 📊 Monitoring

### Prometheus + Grafana

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    volumes:
      - grafana_data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_SERVER_ROOT_URL=http://grafana.yourdomain.com

volumes:
  prometheus_data:
  grafana_data:
```

---

## 🔄 CI/CD Pipeline

### GitHub Actions (Included)

Automatically runs on push:
- ✅ Tests with PostgreSQL
- ✅ Linting
- ✅ Build binaries
- ✅ Build Docker image
- ✅ Upload coverage

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  image: golang:1.21
  services:
    - postgres:15-alpine
  script:
    - go test -v ./...

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

deploy:
  stage: deploy
  script:
    - kubectl set image deployment/pricemap-server server=$CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  only:
    - main
```

---

## 📦 Backup & Recovery

### Database Backup

```bash
# Automated backup script
#!/bin/bash
BACKUP_DIR="/backup/pricemap"
DATE=$(date +%Y%m%d_%H%M%S)

# Backup
docker exec pricemap-db pg_dump -U postgres pricemap > \
  $BACKUP_DIR/pricemap_$DATE.sql

# Compress
gzip $BACKUP_DIR/pricemap_$DATE.sql

# Keep only last 30 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

# Upload to S3 (optional)
aws s3 cp $BACKUP_DIR/pricemap_$DATE.sql.gz \
  s3://your-bucket/backups/
```

Add to cron:

```bash
0 2 * * * /opt/scripts/backup-pricemap.sh
```

### Restore

```bash
# Restore from backup
gunzip < pricemap_20250112_020000.sql.gz | \
  docker exec -i pricemap-db psql -U postgres -d pricemap
```

---

## 🔧 Troubleshooting

### Check Health

```bash
curl http://localhost:3000/health
```

### View Logs

```bash
# Docker
docker-compose logs -f server

# Kubernetes
kubectl logs -f deployment/pricemap-server -n pricemap

# Systemd
journalctl -u pricemap -f
```

### Performance Issues

```bash
# Check resource usage
docker stats pricemap-server

# Database connections
docker exec pricemap-db psql -U postgres -c \
  "SELECT count(*) FROM pg_stat_activity;"

# Increase workers
USE_WORKERS=5 docker-compose up scraper
```

---

## 📞 Support

- 📖 [Full Documentation](COMPREHENSIVE_GUIDE.md)
- 🐛 [Report Issues](https://github.com/VadimToptunov/pricemap-go/issues)
- 💬 [Discussions](https://github.com/VadimToptunov/pricemap-go/discussions)

