# Historical Data API

A production-ready RESTful API service built with Go for managing historical financial data.

## 🚀 Features

- **High Performance**: Fiber v2 framework
- **Clean Architecture**: Clear separation of concerns (Controller → Service → Repository)
- **Database**: MySQL 8.0+ with GORM
- **CSV Upload**: Streaming CSV parser with batch processing (1000 records/batch)
- **Structured Logging**: Zerolog for efficient logging
- **Validation**: Request validation with go-playground/validator
- **Rate Limiting**: IP-based rate limiting
- **Distributed Tracing**: Jaeger with OpenTelemetry
- **Metrics Collection**: Prometheus for HTTP, database, and CSV metrics
- **Visualization**: Grafana dashboards for metrics and traces
- **Log Aggregation**: ELK Stack (Elasticsearch, Logstash, Kibana)
- **Containerization**: Docker & Docker Compose
- **CI/CD**: Complete Jenkins pipeline with automated testing and deployment
- **Production Ready**: Health checks, graceful shutdown, error handling

## 📋 Prerequisites
- Go 1.21 or higher
- Docker & Docker Compose

## 🏗️ Project Structure

```
go-historical-data/
├── cmd/ -- Application entry point
│   └── api/
│       └── main.go
├── config/ -- Configuration files
│   ├── config.dev.yaml
│   ├── config.staging.yaml
│   └── config.prod.yaml
├── database/ -- Database files
│   └── migrations/
├── internal/ -- Private application code
│   ├── controller/
│   ├── dto/
│   │   ├── request/
│   │   └── response/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   └── service/
├── pkg/
│   ├── config/
│   ├── csvparser/
│   ├── database/
│   ├── logger/
│   ├── response/
│   ├── tracing/
│   └── validator/
├── monitoring/ -- Monitoring files
│   ├── elasticsearch/
│   ├── grafana/
│   ├── kibana/
│   └── logstash/
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

## 🚦 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd go-historical-data
```

### 2. Start MySQL and go-historical-data app

```bash
docker-compose up
```

Server will be available:

| Component | URL | Credentials | Purpose |
|-----------|-----|-------------|---------|
| **API** | http://localhost:8080 | None | API Endpoint |    
| **Prometheus** | http://localhost:9090 | None | Metrics collection & querying |
| **Grafana** | http://localhost:3000 | admin / admin | Dashboard visualization |
| **Jaeger UI** | http://localhost:16686 | None | Direct trace analysis |
| **Logstash** | http://localhost:9600 | None | Log parsing and enrichment |
| **Elasticsearch** | http://localhost:9200 | None | Log storage |
| **Kibana (Logs)** | http://localhost:5601 | None | Log visualization |

## 📚 API Endpoints

### Health Check
- `GET /health` - Application health status

### Metrics
- `GET /metrics` - Prometheus metrics endpoint

### Historical Data
- `POST /api/v1/data` - Upload historical data (multipart/form-data)
- `GET /api/v1/data` - Retrieve historical data with filters
- `GET /api/v1/data/:id` - Get specific historical data by ID

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Client                               │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                    Fiber Web Server                          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Middleware: Logging, Tracing, Metrics, Rate Limit  │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                   Controller Layer                           │
│  (HTTP Handlers, Request Validation, Response Formatting)   │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                             │
│  (Business Logic, CSV Processing, Data Transformation)      │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                  Repository Layer                            │
│  (Database Operations, GORM, Query Building)                │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                     MySQL Database                           │
└─────────────────────────────────────────────────────────────┘

                  Observability Stack
┌─────────────────────────────────────────────────────────────┐
│  Prometheus → Grafana (Metrics)                             │
│  Jaeger (Distributed Tracing)                               │
│  ELK Stack (Log Aggregation)                                │
└─────────────────────────────────────────────────────────────┘
```
Test Auto trigger CICD ver 3

**Built with ❤️ using Go, Fiber, MySQL, Docker, Prometheus, Grafana, Jaeger, and ELK Stack**
