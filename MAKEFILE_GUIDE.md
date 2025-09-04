# API Monitor - Makefile Usage Guide

## 🚀 Quick Start

### เริ่มต้นใช้งาน (ครั้งแรก)
```bash
make setup    # ติดตั้ง dependencies และตั้งค่า
make dev      # รัน backend + frontend พร้อมกัน
```

### การใช้งานประจำ
```bash
make dev      # รัน development mode (backend + frontend)
make stop     # หยุดการทำงาน
```

## 📋 Available Commands

### 🔧 Development Commands

| Command | Description | URL |
|---------|-------------|-----|
| `make dev` | รัน backend + frontend พร้อมกัน | Backend: http://localhost:8080<br>Frontend: http://localhost:5173 |
| `make dev-backend` | รัน backend อย่างเดียว | http://localhost:8080 |
| `make dev-frontend` | รัน frontend อย่างเดียว | http://localhost:5173 |

### 🏗️ Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build ทั้ง backend และ frontend |
| `make build-backend` | Build backend binary |
| `make build-frontend` | Build frontend สำหรับ production |

### ⚙️ Setup Commands

| Command | Description |
|---------|-------------|
| `make install` | ติดตั้ง dependencies ทั้งหมด |
| `make setup` | ติดตั้งและตั้งค่าเริ่มต้น |

### 🧹 Maintenance Commands

| Command | Description |
|---------|-------------|
| `make clean` | ลบไฟล์ build artifacts |
| `make test` | รัน tests |
| `make stop` | หยุดการทำงานทั้งหมด |

### 📊 Utility Commands

| Command | Description |
|---------|-------------|
| `make info` | แสดงข้อมูล environment |
| `make health` | ตรวจสอบสถานะ backend/frontend |
| `make restart` | หยุดและเริ่มใหม่ |

## 🔄 Typical Workflow

### การพัฒนาปกติ:
```bash
# เริ่มต้น
make dev

# หยุดเมื่อเสร็จงาน
make stop
```

### การ setup ใหม่:
```bash
# ติดตั้งครั้งแรก
make setup

# ตรวจสอบสถานะ
make info

# รันทดสอบ
make dev
```

### การแก้ไข backend อย่างเดียว:
```bash
# รัน backend อย่างเดียว
make dev-backend
```

### การแก้ไข frontend อย่างเดียว:
```bash
# รัน frontend อย่างเดียว  
make dev-frontend
```

## ⚠️ Prerequisites

ต้องติดตั้งก่อนใช้งาน:
- **Go** 1.19+ 
- **Node.js** 16+
- **NPM** 8+

ตรวจสอบด้วย:
```bash
make info
```

## 🔧 Configuration

### Backend (.env required):
```bash
DB_HOST=your-database-host
DB_PORT=5432
DB_USERNAME=your-username
DB_PASSWORD=your-password
DB_DATABASE=your-database
JWT_SECRET=your-jwt-secret
```

### Frontend (auto-configured):
- Development: `http://localhost:8080/api/v1`
- Production: ตั้งค่าใน `.env.production`

## 🐛 Troubleshooting

### Port Already in Use:
```bash
make stop    # หยุดกระบวนการที่รันอยู่
make dev     # เริ่มใหม่
```

### Dependencies Issues:
```bash
make clean   # ลบไฟล์ build
make install # ติดตั้ง dependencies ใหม่
```

### Database Connection Issues:
1. ตรวจสอบ `.env` file ใน backend directory
2. ตรวจสอบการเชื่อมต่อ database
3. รัน `make dev-backend` เพื่อดู error logs

### Frontend Build Issues:
```bash
cd frontend
rm -rf node_modules package-lock.json
make install
```

## 🎯 Examples

### Development Session:
```bash
# เช้า - เริ่มงาน
make dev

# กลางวัน - ตรวจสอบสถานะ
make health

# เย็น - หยุดงาน
make stop
```

### Production Build:
```bash
make build
# ไฟล์จะอยู่ที่:
# - backend/api-monitor (binary)
# - frontend/dist/ (static files)
```

### Testing Changes:
```bash
make test     # รัน tests
make restart  # restart services
make health   # ตรวจสอบว่าทำงานปกติ
```

## 🚀 Performance Tips

- ใช้ `make dev-backend` หรือ `make dev-frontend` เมื่อแก้ไขเฉพาะด้านใดด้านหนึ่ง
- ใช้ `make restart` แทน `make stop && make dev` เพื่อความเร็ว
- ใช้ `make health` เพื่อตรวจสอบสถานะเป็นระยะ
