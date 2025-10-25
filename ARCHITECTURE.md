# URL Shortener System Architecture

## High-Level Design

## Component Responsibilities

### 1. Application Layer
- **Stateless Go application** with hexagonal architecture
- **HTTP API** for URL shortening and redirection
- **Business logic** for short code generation and validation
- **Caching integration** for performance optimization

### 2. Data Layer
- **PostgreSQL** - Primary data store with ACID compliance
- **Redis** - Distributed caching for high-read throughput
- **Connection pooling** for database efficiency

### 3. Infrastructure Layer
- **Nginx** - Load balancing and SSL termination
- **Docker** - Containerization for consistent deployments
- **Health checks** - For service discovery and load balancing

## Data Flow

### URL Shortening Flow
1. Client POSTs long URL to `/api/shorten`
2. Application validates URL format and length
3. Generate unique short code (base62, 6 characters)
4. Store URL mapping in PostgreSQL
5. Return short URL to client

### URL Redirection Flow
1. Client GETs short URL `/{code}`
2. Check Redis cache first (cache hit → return immediately)
3. Cache miss: Query PostgreSQL for original URL
4. Cache the result in Redis for future requests
5. Return 301 redirect to original URL
6. Async increment click counter

## Technology Stack

**Go 1.24+** - Excellent performance with built-in concurrency and minimal runtime overhead

**Standard Library** - No framework lock-in, maximum control over HTTP handling, and minimal dependencies

**Hexagonal Architecture** - Clean separation of concerns, highly testable, and maintainable codebase

**PostgreSQL** - ACID compliance for data integrity, JSON support for flexibility, and excellent read/write performance

**Redis** - High-performance caching layer with persistence options and atomic operations for click counting

**Docker** - Consistent development and deployment environments, cloud-agnostic containerization

## Key Design Decisions

### 1. Short Code Generation
- **Base62 encoding** (a-z, A-Z, 0-9) for compact URLs
- **6 characters** provides 56.8 billion possible combinations
- **Cryptographically random** generation to prevent prediction
- **Collision handling** with retry logic

### 2. Caching Strategy
- **Write-through cache** for new URLs
- **1-hour TTL** balances freshness and performance
- **LRU eviction** when memory limits reached
- **Async cache population** for redirects


## ⚖️ Trade-offs and Assumptions

**File: `TRADE-OFFS.md`**
```markdown
# Design Trade-offs and Assumptions

## Explicit Trade-offs

### 1. Simplicity vs. Features
**Choice**: Started with core functionality only
**Trade-off**: No user accounts, analytics, or custom URLs initially
**Rationale**: Meet assessment requirements without over-engineering
**Future**: Can add features incrementally

### 2. Performance vs. Consistency
**Choice**: Eventual consistency for click counts
**Trade-off**: Click counts may be slightly delayed
**Rationale**: Redirect performance is critical path
**Future**: Can implement stronger consistency if needed

### 3. Memory vs. Performance
**Choice**: In-memory Redis cache
**Trade-off**: Potential data loss on cache failure
**Rationale**: URLs are persisted in PostgreSQL, cache is for performance
**Future**: Can enable Redis persistence

### 4. Code Complexity vs. Observability
**Choice**: Lightweight custom metrics vs. Prometheus
**Trade-off**: Less sophisticated monitoring capabilities
**Rationale**: Avoid heavy dependencies for assessment
**Future**: Can integrate Prometheus later

## Technical Assumptions

### 1. Traffic Patterns
- **Read-heavy workload**: 80% redirects, 20% shortening
- **Peak traffic**: 10,000 requests per second
- **URL distribution**: Power-law (some URLs very popular)

### 2. Data Characteristics
- **URL length**: Average 100 characters, max 2048
- **Short code length**: 6 characters fixed
- **Data growth**: ~1 million URLs per month

### 3. Infrastructure Assumptions
- **Network latency**: <10ms between services
- **Database**: SSD storage, sufficient RAM for working set
- **Cache**: Enough memory for hot URL dataset

## Limitations and Constraints

### 1. Current Limitations
- No user authentication or authorization
- No URL expiration or manual deletion
- No advanced analytics or reporting
- No bulk operations or import/export

### 2. Scalability Boundaries
- **Single PostgreSQL**: Up to ~50,000 RPS with proper indexing
- **Single Redis**: Up to ~100,000 RPS for cached reads
- **Application**: Stateless, scales horizontally easily

### 3. Security Considerations
- No rate limiting per user (only per IP)
- No malicious URL detection
- No abuse prevention beyond basic rate limiting

## Future Enhancement Path

### Phase 1 (Current)
- Core shortening and redirection
- Basic caching and observability

### Phase 2 (Next)
- User accounts and URL management
- Custom short codes
- URL expiration

### Phase 3 (Advanced)
- Advanced analytics
- API rate limiting per user
- Bulk operations
- Admin dashboard