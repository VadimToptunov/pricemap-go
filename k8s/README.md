# Kubernetes Deployment

Deploy PriceMap-Go to Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (Minikube, GKE, EKS, AKS, etc.)
- `kubectl` configured
- Docker image built and available

## Quick Start

### 1. Build and Push Docker Image

```bash
# Build image
docker build -t pricemap-go:latest .

# Tag for registry (if using remote cluster)
docker tag pricemap-go:latest your-registry/pricemap-go:latest

# Push to registry
docker push your-registry/pricemap-go:latest
```

### 2. Update Secrets

Edit `deployment.yaml` and change default password:

```yaml
stringData:
  DB_PASSWORD: "your-secure-password-here"  # Change this!
```

### 3. Deploy

```bash
# Apply all resources
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get pods -n pricemap
kubectl get services -n pricemap
```

### 4. Access Application

```bash
# Get service URL (for LoadBalancer)
kubectl get svc pricemap-server-service -n pricemap

# Or port-forward for testing
kubectl port-forward -n pricemap svc/pricemap-server-service 3000:80
```

## Components

### Deployments

- **PostgreSQL** - 1 replica with persistent storage
- **Tor** - 2 replicas for load balancing
- **Server** - 3 replicas with auto-scaling (2-10)

### Services

- **postgres-service** - Internal database access
- **tor-service** - Internal Tor proxy access
- **pricemap-server-service** - External HTTP access (LoadBalancer)

### Jobs

- **pricemap-scraper** - CronJob runs every 6 hours

### Storage

- **postgres-pvc** - 10Gi persistent volume for database

## Configuration

### ConfigMap

Environment variables in `pricemap-config`:
- Server settings
- Tor configuration
- Rate limiting

### Secrets

Sensitive data in `pricemap-db-secret`:
- Database credentials

## Scaling

### Manual Scaling

```bash
# Scale server
kubectl scale deployment pricemap-server -n pricemap --replicas=5

# Scale Tor
kubectl scale deployment tor -n pricemap --replicas=3
```

### Auto-scaling

HorizontalPodAutoscaler configured:
- Min: 2 replicas
- Max: 10 replicas
- Triggers: CPU > 70%, Memory > 80%

## Monitoring

### Health Checks

```bash
# Check liveness
kubectl exec -it -n pricemap <pod-name> -- wget -O- http://localhost:3000/liveness

# Check readiness
kubectl exec -it -n pricemap <pod-name> -- wget -O- http://localhost:3000/readiness

# Full health
kubectl exec -it -n pricemap <pod-name> -- wget -O- http://localhost:3000/health
```

### Logs

```bash
# Server logs
kubectl logs -f -n pricemap deployment/pricemap-server

# Scraper logs
kubectl logs -n pricemap job/pricemap-scraper-<timestamp>

# Database logs
kubectl logs -f -n pricemap deployment/postgres
```

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl describe pod -n pricemap <pod-name>

# Check events
kubectl get events -n pricemap --sort-by='.lastTimestamp'
```

### Database connection issues

```bash
# Test database connectivity
kubectl run -it --rm debug -n pricemap --image=postgres:15-alpine --restart=Never -- \
  psql -h postgres-service -U postgres -d pricemap
```

### Tor issues

```bash
# Test Tor proxy
kubectl exec -it -n pricemap deployment/tor -- \
  curl --socks5 localhost:9050 https://check.torproject.org/api/ip
```

## Production Recommendations

1. **Use Secrets Management**: Consider Vault or Sealed Secrets
2. **Enable TLS**: Use cert-manager for automatic SSL certificates
3. **Setup Monitoring**: Deploy Prometheus and Grafana
4. **Configure Backup**: Automate PostgreSQL backups
5. **Resource Limits**: Adjust based on actual usage
6. **Network Policies**: Restrict inter-pod communication
7. **Pod Disruption Budgets**: Ensure availability during updates

## Cleanup

```bash
# Delete all resources
kubectl delete -f k8s/deployment.yaml

# Or delete namespace
kubectl delete namespace pricemap
```

## Advanced Configuration

### Using Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: pricemap-ingress
  namespace: pricemap
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - pricemap.yourdomain.com
    secretName: pricemap-tls
  rules:
  - host: pricemap.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: pricemap-server-service
            port:
              number: 80
```

### Resource Quotas

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: pricemap-quota
  namespace: pricemap
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 4Gi
    limits.cpu: "8"
    limits.memory: 8Gi
    persistentvolumeclaims: "5"
```

## Support

For issues and questions, see main [README.md](../README.md)

