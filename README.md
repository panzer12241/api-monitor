# API Monitor

ระบบตรวจสอบสถานะ API แบบ real-time ที่สามารถ monitor API endpoints ต่าง ๆ และแสดงผลผ่าน Grafana dashboard

## Features

- ✅ เพิ่ม/ลบ/แก้ไข API endpoints ผ่าน Web UI
- ✅ กำหนดช่วงเวลาการตรวจสอบ (วินาที/นาที/ชั่วโมง)
- ✅ เปิด/ปิดการตรวจสอบแต่ละ endpoint
- ✅ ดู log ย้อนหลัง (status code, response time, error messages)
- ✅ Real-time dashboard
- ✅ Dockerized deployment

## Technology Stack

- **Backend**: Go (Gin framework)
- **Frontend**: Vue.js 3 + Element Plus + Vite + Bun
- **Database**: PostgreSQL (configuration)
- **Scheduler**: Go goroutines with cron
- **Deployment**: Docker & Docker Compose

## Quick Start

### 1. Clone และเข้าไปในโฟลเดอร์

```bash
cd api-monitor
```

### 2. รัน Docker Compose

```bash
docker-compose up -d
```

### 3. เข้าใช้งาน

- **Web UI**: https://monitor.maxnano.app
- **Backend API**: https://monitor-api.maxnano.app

## Project Structure

```
api-monitor/
├── backend/                 # Go API backend
│   ├── main.go             # Main application
│   ├── Dockerfile          # Backend Docker image
│   └── go.mod              # Go dependencies
├── frontend/               # Vue.js frontend
│   ├── src/
│   │   ├── views/
│   │   │   ├── Dashboard.vue    # Dashboard page
│   │   │   └── Endpoints.vue    # Endpoint management
│   │   ├── App.vue         # Main app component
│   │   └── main.js         # App entry point
│   ├── Dockerfile          # Frontend Docker image
│   └── package.json        # NPM dependencies
├── docker/                 # Docker configurations
│   └── postgres/           # PostgreSQL init scripts
└── docker-compose.yml      # Main orchestration file
```

## API Endpoints

### Endpoint Management
- `GET /api/v1/endpoints` - รายการ endpoints ทั้งหมด
- `POST /api/v1/endpoints` - สร้าง endpoint ใหม่
- `PUT /api/v1/endpoints/:id` - แก้ไข endpoint
- `DELETE /api/v1/endpoints/:id` - ลบ endpoint
- `POST /api/v1/endpoints/:id/toggle` - เปิด/ปิดการตรวจสอบ
- `POST /api/v1/endpoints/:id/check` - ตรวจสอบ manual
- `GET /api/v1/endpoints/:id/logs` - ดู logs

## Environment Variables

### Backend
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres123)
- `DB_NAME` - Database name (default: api_monitor)

## Development

### Prerequisites
- Docker & Docker Compose
- Bun (for frontend development)

### Backend Development

```bash
cd backend
go mod tidy
go run main.go
```

### Frontend Development with Bun

```bash
# Install Bun if not already installed
curl -fsSL https://bun.sh/install | bash

# Start frontend development server
./dev-frontend.sh

# Or manually:
cd frontend
bun install
bun run dev
```

### Database Schema

```sql
-- API endpoints configuration
CREATE TABLE api_endpoints (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    method VARCHAR(10) DEFAULT 'GET',
    headers JSONB DEFAULT '{}',
    body TEXT,
    timeout_seconds INTEGER DEFAULT 30,
    check_interval_seconds INTEGER DEFAULT 60,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Check logs
CREATE TABLE api_check_logs (
    id SERIAL PRIMARY KEY,
    endpoint_id INTEGER REFERENCES api_endpoints(id),
    status_code INTEGER,
    response_time_ms INTEGER,
    response_body TEXT,
    error_message TEXT,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Configuration Examples

### สร้าง API Endpoint

```json
{
  "name": "My API Health Check",
  "url": "https://api.example.com/health",
  "method": "GET",
  "headers": {
    "Authorization": "Bearer token123",
    "Content-Type": "application/json"
  },
  "timeout_seconds": 30,
  "check_interval_seconds": 60,
  "is_active": true
}
```

### HTTP Methods Support
- GET, POST, PUT, DELETE, PATCH, HEAD
- รองรับ custom headers
- รองรับ request body สำหรับ POST/PUT requests

## Monitoring Capabilities

1. **Real-time Status**: ตรวจสอบสถานะ API แบบ real-time
2. **Historical Data**: เก็บ logs ย้อนหลังใน PostgreSQL
3. **Custom Intervals**: กำหนดช่วงเวลาตรวจสอบแต่ละ endpoint ได้

## Troubleshooting

### ตรวจสอบ container status
```bash
docker-compose ps
```

### ดู logs
```bash
docker-compose logs api-monitor-backend
docker-compose logs api-monitor-frontend
```

### Restart services
```bash
docker-compose restart
```

## License

MIT License