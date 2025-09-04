# API Monitor Backend

## Laravel-like Project Structure

```
backend/
├── app/
│   ├── controllers/        # HTTP Controllers (เหมือน Laravel Controllers)
│   │   ├── auth_controller.go
│   │   └── endpoint_controller.go
│   ├── middleware/         # HTTP Middleware (เหมือน Laravel Middleware)
│   │   └── auth.go
│   ├── models/            # Data Models (เหมือน Laravel Models)
│   │   ├── user.go
│   │   ├── endpoint.go
│   │   └── proxy.go
│   └── services/          # Business Logic Services (เหมือน Laravel Services)
│       └── monitor.go
├── cmd/                   # Application Entry Points
│   └── main.go
├── config/                # Configuration Files (เหมือน Laravel Config)
│   └── database.go
├── database/              # Database Related Files (เหมือน Laravel Database)
│   └── migrations/        # Database Migrations (เหมือน Laravel Migrations)
│       ├── auth-migration.sql
│       ├── migration.sql
│       └── proxy-migration.sql
├── routes/                # Route Definitions (เหมือน Laravel Routes)
│   └── api.go
├── utils/                 # Utility Functions (เหมือน Laravel Helpers)
│   └── http.go
├── .env                   # Environment Variables
├── .env.example          # Environment Variables Example
├── go.mod                # Go Module Dependencies
├── go.sum                # Go Module Checksums
└── README.md             # This file
```

## การเปลี่ยนแปลงจากโครงสร้างเดิม

### ก่อน (Monolithic)
```
backend/
├── main.go              # ทุกอย่างอยู่ในไฟล์เดียว (1,287 บรรทัด)
├── migration.sql
├── auth-migration.sql
├── proxy-migration.sql
├── .env
├── go.mod
└── go.sum
```

### หลัง (Laravel-like Structure)
- แยกโค้ดออกเป็น modules ตาม responsibility
- ง่ายต่อการ maintain และ scale
- ง่ายต่อการ test แต่ละส่วน
- ง่ายต่อการทำงานเป็นทีม

## วิธีการ Build และ Run

```bash
# Build
go build -o api-monitor ./cmd/main.go

# Run
./api-monitor

# หรือ Run โดยตรง
go run ./cmd/main.go
```

## Features

### App Structure
- **Controllers**: จัดการ HTTP requests และ responses
- **Models**: ข้อมูล structs และ business models
- **Middleware**: Authentication, CORS, และ middleware อื่นๆ
- **Services**: Business logic และ background processes
- **Routes**: จัดการ routing และ middleware chains
- **Utils**: Helper functions และ utilities
- **Config**: การตั้งค่าและ configuration management

### Laravel-like Features Implemented
1. **MVC Pattern**: Model-View-Controller architecture
2. **Middleware**: Authentication และ authorization
3. **Service Layer**: Business logic separation
4. **Configuration Management**: Environment-based config
5. **Database Migrations**: Structured database changes
6. **Routing**: Clean route definitions with middleware
7. **Dependency Injection**: Service initialization

## Environment Variables

```bash
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=api_monitor
DB_USERNAME=postgres
DB_PASSWORD=your_password
JWT_SECRET=your_jwt_secret
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration

### Endpoints Management (Requires JWT)
- `GET /api/v1/endpoints` - Get all endpoints
- `POST /api/v1/endpoints` - Create endpoint
- `PUT /api/v1/endpoints/:id` - Update endpoint
- `DELETE /api/v1/endpoints/:id` - Delete endpoint
- `POST /api/v1/endpoints/:id/toggle` - Toggle endpoint status

### User Management (Admin only)
- `GET /api/v1/users` - Get all users
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## Technologies Used

- **Fiber**: Fast HTTP framework for Go
- **PostgreSQL**: Database
- **JWT**: Authentication
- **Cron**: Scheduled tasks
- **Bcrypt**: Password hashing
