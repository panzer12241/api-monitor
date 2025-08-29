#!/bin/bash

echo "🚀 Starting API Monitor Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Build and start all services
echo "📦 Building and starting services..."
docker-compose up --build -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 30

# Check service health
echo "🔍 Checking service health..."

# Check PostgreSQL
echo "   Checking PostgreSQL..."
docker-compose exec -T postgres pg_isready -U postgres

# Check Backend API
echo "   Checking Backend API..."
curl -f http://localhost:8080/metrics > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✅ Backend API is ready"
else
    echo "   ⚠️  Backend API is not ready yet"
fi

# Check Frontend
echo "   Checking Frontend..."
curl -f http://localhost:8081 > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✅ Frontend is ready"
else
    echo "   ⚠️  Frontend is not ready yet"
fi

# Check Grafana
echo "   Checking Grafana..."
curl -f http://localhost:3000 > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✅ Grafana is ready"
else
    echo "   ⚠️  Grafana is not ready yet"
fi

# Check Prometheus
echo "   Checking Prometheus..."
curl -f http://localhost:9090 > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "   ✅ Prometheus is ready"
else
    echo "   ⚠️  Prometheus is not ready yet"
fi

echo ""
echo "🎉 API Monitor is running!"
echo ""
echo "📱 Access URLs:"
echo "   Frontend:   http://localhost:8081"
echo "   Grafana:    http://localhost:3000 (admin/admin123)"
echo "   Prometheus: http://localhost:9090"
echo "   Backend:    http://localhost:8080"
echo ""
echo "🛠️  Useful commands:"
echo "   Stop:       docker-compose down"
echo "   Logs:       docker-compose logs -f"
echo "   Restart:    docker-compose restart"
echo ""
