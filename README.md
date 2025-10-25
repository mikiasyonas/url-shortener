# URL Shortener Service
A high-performance, production-ready URL shortening service built with Go and Hexagonal Architecture. 
Capable of handling 10,000+ requests per second with proper scaling.

## Features
**High Performance**: Redis caching, connection pooling, and optimized database queries

**Scalable Architecture**: Horizontal scaling ready with load balancing

**Production Ready**: Comprehensive observability, health checks, and monitoring

**Simple API**: RESTful endpoints for URL shortening and redirection

**Containerized**: Docker and Docker Compose for easy deployment

# 🏗️ Architecture
## Technology Stack
### Component	Technology	Justification
**Language**	Go 1.21+	Excellent performance, built-in concurrency, minimal runtime
**Architecture**	Hexagonal/Ports & Adapters	Clean separation, testable, maintainable
**Database**	PostgreSQL	ACID compliance, JSON support, excellent performance
**Caching**	Redis	High-performance, persistence options, atomic operations
**Containerization**	Docker	Consistent environments, cloud-agnostic deployment


## System Design

┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client        │    │   Load Balancer  │    │   App Servers   │
│                 │───▶│   (Nginx)        │───▶│   (4+ replicas) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                         │
                                                         ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Monitoring    │◀───│   Redis Cache    │◀───│   PostgreSQL    │
│   (Metrics)     │    │   (Cluster)      │    │   (Primary +    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                          │
                                                          ▼
                                                   ┌─────────────────┐
                                                   │   Read Replicas │
                                                   │   (2 instances) │
                                                   └─────────────────┘

# Quick Start
## Prerequisites

### Docker and Docker Compose
2GB RAM minimum, 4GB recommended

### Development Setup
Clone and setup environment

bash
git clone <repository-url>
cd url-shortener
cp .env.example .env

### Start services

bash
# Start database and Redis
docker-compose up -d postgres redis

# Wait for databases to be ready
sleep 10

# Start application
docker-compose up -d app

# Check status
docker-compose ps
Verify deployment

bash
# Health check
curl http://localhost:8080/health

# Test URL shortening
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'

# Test redirection (replace abc123 with actual short code)
curl -I http://localhost:8080/abc123
🚀 Production Deployment
1. Production Environment Setup
bash
cp .env.production.example .env.production
nano .env.production  # Edit with your production values
Sample production configuration:

env
ENVIRONMENT=production
APP_BASE_URL=https://short.yourdomain.com
DB_PASSWORD=your-strong-password
REDIS_PASSWORD=your-redis-password
2. Deploy Production Stack
bash
# Use production compose file
docker-compose -f deployments/docker-compose.prod.yml --env-file .env.production up -d

# Scale application instances
docker-compose -f deployments/docker-compose.prod.yml up -d --scale app=4

# Monitor deployment
docker-compose -f deployments/docker-compose.prod.yml logs -f
3. Cloud Deployment (AWS Example)
bash
# Build and push to ECR
docker build -f deployments/Dockerfile -t url-shortener:latest .
docker tag url-shortener:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/url-shortener:latest
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/url-shortener:latest

# Deploy to ECS
aws ecs register-task-definition --cli-input-json file://deployments/aws/task-definition.json
aws ecs update-service --cluster production --service url-shortener --task-definition url-shortener
📚 API Usage
Shorten URL
Endpoint: POST /api/shorten

Request:

json
{
  "url": "https://example.com/very/long/url/path/that/needs/shortening"
}
Response:

json
{
  "success": true,
  "data": {
    "short_url": "http://localhost:8080/abc123",
    "original_url": "https://example.com/very/long/url/path/that/needs/shortening",
    "short_code": "abc123"
  }
}
Redirect to Original URL
Endpoint: GET /{shortCode}

Response:

301 Moved Permanently redirect to original URL

Health Checks
GET /health - Comprehensive health status

GET /ready - Readiness for load balancers

GET /live - Liveness check

GET /metrics - Application metrics