# Scaling Strategy for URL Shortener Service

---

## Current Capacity

### Baseline Performance (Single Instance)

* **Redirects:** 5,000–8,000 RPS (with Redis cache)
* **URL Shortening:** 500–1,000 RPS
* **Memory Usage:** ~200MB per instance
* **Database Connections:** 25–50 active

**Target Capacity:** 10,000 RPS

---

## Horizontal Scaling Plan

### Step 1: Application Layer Scaling

**Configuration:**

```yaml
# docker-compose.prod.yml
app:
  deploy:
    replicas: 4
    resources:
      limits:
        memory: 512M
        cpu: '0.5'
```

**Load Balancer Configuration:**

```nginx
# Nginx configuration
upstream app_servers {
    least_conn;
    server app1:8080 max_fails=3 fail_timeout=30s;
    server app2:8080 max_fails=3 fail_timeout=30s;
    server app3:8080 max_fails=3 fail_timeout=30s;
    server app4:8080 max_fails=3 fail_timeout=30s;
}
```

---

### Step 2: Database Scaling

**Read Replicas:**

* **Primary:** Handles all writes (URL shortening)
* **Replica 1:** Handles 50% of redirect reads
* **Replica 2:** Handles 50% of redirect reads

**Connection Pool Settings:**

```env
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=5m
```

---

### Step 3: Cache Scaling

**Redis Cluster:**

```yaml
# Redis cluster with 3 nodes
redis:
  image: redis:7-alpine
  command: redis-server --cluster-enabled yes --cluster-config-file nodes.conf --cluster-node-timeout 5000
  deploy:
    replicas: 3
```

**Cache Configuration:**

```env
REDIS_POOL_SIZE=100
REDIS_TTL=1h
```

---

## Auto-scaling Configuration

### Kubernetes HPA

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: url-shortener
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: url-shortener
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: 2000
```

### Cloud Auto-scaling (AWS)

```json
{
  "TargetTrackingPolicies": [
    {
      "PredefinedMetricSpecification": {
        "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
      },
      "TargetValue": 70.0
    },
    {
      "PredefinedMetricSpecification": {
        "PredefinedMetricType": "ECSServiceAverageMemoryUtilization"
      },
      "TargetValue": 80.0
    }
  ]
}
```

---

## Performance Monitoring

### Key Metrics to Monitor

#### Application Metrics

* Request rate per second
* Response time (p95, p99)
* Error rate (4xx, 5xx)
* Cache hit rate
* Active connections

#### Database Metrics

* Query throughput
* Replication lag
* Connection pool usage
* Lock contention

#### Infrastructure Metrics

* CPU utilization
* Memory usage
* Network I/O
* Disk I/O

---

### Alerting Thresholds

#### Critical Alerts

* Error rate > **1%** for 2 minutes
* P95 response time > **500ms** for 5 minutes
* Database connections > **90%** for 2 minutes

#### Warning Alerts

* Cache hit rate < **80%** for 10 minutes
* CPU utilization > **70%** for 5 minutes
* Memory usage > **80%** for 5 minutes

---

## Load Testing

### Test Scenarios

#### Baseline Test

```bash
# 1,000 RPS for 5 minutes
k6 run --vus 1000 --duration 5m loadtest/redirects.js
```

#### Peak Load Test

```bash
# 10,000 RPS for 10 minutes
k6 run --vus 10000 --duration 10m loadtest/peak.js
```

#### Stress Test

```bash
# Gradually increase to 20,000 RPS
k6 run --vus 20000 --duration 15m loadtest/stress.js
```

---

### Expected Results

At **10,000 RPS**:

* **P95 latency (redirects):** < 100ms
* **P95 latency (shortening):** < 500ms
* **Error rate:** < 0.1%
* **Cache hit rate:** > 85%

---

## Cost Optimization

### Resource Right-sizing

**Development:**

* 2 app instances, 1GB RAM each
* Single PostgreSQL instance, 4GB RAM
* Single Redis instance, 1GB RAM

**Production:**

* 4–8 app instances, 512MB RAM each
* PostgreSQL with read replicas
* Redis cluster with 3 nodes

---

## Disaster Recovery

### Backup Strategy

**Database:**

* Automated daily backups
* Point-in-time recovery enabled
* Cross-region replication

**Application:**

* Container images in registry
* Infrastructure as Code
* Configuration in version control

---

### Recovery Procedures

**Database Failure:**

* Promote read replica to primary
* Update application configuration
* Restore from backup if needed

**Cache Failure:**

* Application continues with database fallback
* Gradually repopulate cache
* Monitor performance during recovery

**Application Failure:**

* Load balancer detects unhealthy instances
* Auto-scaling launches new instances
* Traffic routed to healthy instances

---

## Performance Optimizations Implemented

### 1. Database Layer

* Connection pooling (25–50 connections)
* Proper indexing on `short_code` and `created_at`
* Read replicas ready for horizontal scaling
* Query optimization

### 2. Caching Layer

* Redis caching for 80–90% of redirect requests
* 1-hour TTL for cached URLs
* Connection pooling (100 connections)
* LRU eviction policy

### 3. Application Layer

* Stateless application servers
* Goroutine-per-request model
* Connection reuse with Keep-Alive
* Efficient short code generation

### 4. Infrastructure

* Load balancing with Nginx
* Horizontal scaling ready
* Health checks and graceful shutdown

---

## Future Scaling Considerations

### For 50,000+ RPS

* Implement database sharding by short code prefix
* Add message queue for async operations
* Implement regional caching

### For 100,000+ RPS

* Geographic distribution with multiple regions
* Database partitioning and advanced replication
* Advanced load balancing with anycast routing