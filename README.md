# Order Processing System with Temporal.io

A simplified order processing system built with Go, Temporal.io, and PostgreSQL.

## 🚀 Features

- **Temporal Workflows**: Orchestrates order processing steps
- **PostgreSQL Integration**: Persistent data storage
- **Docker Support**: Easy deployment with Docker Compose
- **Web Interface**: Simple UI for order management
- **Error Handling**: Comprehensive error handling with retries
- **Activity Retries**: Configurable retry policies for failed activities

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Interface │    │   API Server    │    │  Temporal      │
│   (Port 8080)   │◄──►│   (Port 8080)   │◄──►│  (Port 7233)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   PostgreSQL    │    │     Redis       │
                       │   (Port 5432)   │    │   (Port 6379)   │
                       └─────────────────┘    └─────────────────┘
```

## 📋 Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Make (optional, for convenience)

## 🐳 Quick Start with Docker

### 1. Start the full stack

```bash
make start-full-stack
```

This will start:

- PostgreSQL database on port 5432
- Redis on port 6379
- Temporal server on port 7233
- Temporal Web UI on port 8088
- Your application API on port 8080

### 2. Check services

```bash
# View all running containers
docker-compose ps

# View logs
make docker-logs

# Check specific service logs
docker-compose logs -f api
```

### 3. Access services

- **Web Interface**: http://localhost:8080
- **Temporal Web UI**: http://localhost:8088
- **PostgreSQL**: localhost:5432 (user: orderflow, password: orderflow123)

## 🛠️ Development

### Local Development

```bash
# Install dependencies
make deps

# Build the application
make build

# Run worker
make run-worker

# Run API server
make run-api

# Run both
make run
```

### Docker Development

```bash
# Build containers
make docker-build

# Start services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down

# Clean everything
make docker-clean
```

## 🗄️ Database

### Schema

The system includes the following tables:

- `orders` - Main order information
- `order_items` - Individual items in orders
- `products` - Product catalog
- `payments` - Payment records
- `notifications` - Customer notifications

### Sample Data

Sample products are automatically inserted:

- PROD-001: Laptop Dell XPS 13 ($1,299.99)
- PROD-002: iPhone 15 Pro ($999.99)
- PROD-003: Samsung 4K TV ($799.99)
- PROD-004: Wireless Headphones ($299.99)
- PROD-005: Gaming Mouse ($89.99)

### Database Commands

```bash
# Connect to database
make db-connect

# Reset database (careful - deletes all data)
make db-reset

# Run migrations manually
make db-migrate
```

## 🔧 Configuration

### Environment Variables

- `POSTGRES_HOST` - PostgreSQL host (default: localhost)
- `POSTGRES_PORT` - PostgreSQL port (default: 5432)
- `POSTGRES_USER` - Database user (default: orderflow)
- `POSTGRES_PASSWORD` - Database password (default: orderflow123)
- `POSTGRES_DB` - Database name (default: orderflow)
- `TEMPORAL_HOST_PORT` - Temporal server address (default: localhost:7233)

### Docker Environment

All environment variables are configured in `docker-compose.yml` and automatically passed to containers.

## 📊 API Endpoints

### Create Order

```bash
POST /api/orders
Content-Type: application/json

{
  "customer_id": "CUST-001",
  "items": [
    {
      "product_id": "PROD-001",
      "product_name": "Laptop Dell XPS 13",
      "quantity": 1,
      "unit_price": 1299.99
    }
  ]
}
```

### Check Order Status

```bash
GET /api/orders/{orderID}/status
```

### Cancel Order

```bash
POST /api/orders/{orderID}/cancel
```

## 🔄 Workflow Steps

1. **Order Creation** - Creates order record in database
2. **Inventory Check** - Verifies product availability
3. **Payment Processing** - Processes payment transaction
4. **Customer Notification** - Sends confirmation email/SMS

## ❌ Error Handling

- **Retry Policies**: Activities automatically retry on failure
- **Non-Retryable Errors**: Certain errors (e.g., invalid product) don't retry
- **Error Logging**: All errors are logged with context
- **Status Updates**: Order status reflects current processing state

## 🚫 Cancellation Support

Orders can be cancelled using Temporal Signals:

- Send cancellation signal via API
- Workflow gracefully handles cancellation
- Updates order status to cancelled

## 🔍 Monitoring

### Temporal Web UI

- View workflow executions
- Monitor activity status
- Debug failed workflows
- Access: http://localhost:8088

### Application Logs

```bash
# View API logs
docker-compose logs -f api

# View worker logs
docker-compose logs -f worker

# View all logs
make docker-logs
```

## 🧪 Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint
```

## 🚀 Production Deployment

### Docker Production

```bash
# Build production images
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

# Deploy with production config
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Environment Variables

Set production environment variables:

- Database credentials
- Temporal cluster endpoints
- Logging levels
- API keys for external services

## 🐛 Troubleshooting

### Common Issues

#### 1. Port Already in Use

```bash
# Check what's using the port
netstat -tulpn | grep :8080

# Kill process or change port in docker-compose.yml
```

#### 2. Database Connection Issues

```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Verify database is running
docker-compose exec postgres pg_isready -U orderflow
```

#### 3. Temporal Connection Issues

```bash
# Check Temporal logs
docker-compose logs temporal

# Verify Temporal is ready
curl http://localhost:8233/health
```

#### 4. Container Build Issues

```bash
# Clean build cache
docker system prune -a

# Rebuild without cache
docker-compose build --no-cache
```

### Reset Everything

```bash
# Stop and remove all containers, volumes, and images
make docker-clean

# Start fresh
make start-full-stack
```

## 📚 Additional Resources

- [Temporal.io Documentation](https://docs.temporal.io/)
- [Go Temporal SDK](https://docs.temporal.io/go)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.
