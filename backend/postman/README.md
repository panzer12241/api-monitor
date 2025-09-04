# Postman Collection - API Monitor Backend

## 📁 ไฟล์ที่จำเป็น

1. **API-Monitor-Backend.postman_collection.json** - Collection หลัก
2. **API-Monitor-Environment.postman_environment.json** - Environment variables

## 🚀 การตั้งค่า

### 1. Import ไฟล์เข้า Postman

1. เปิด Postman
2. คลิก **Import** 
3. เลือกไฟล์ทั้งสอง:
   - `API-Monitor-Backend.postman_collection.json`
   - `API-Monitor-Environment.postman_environment.json`

### 2. เลือก Environment

1. ที่มุมขวาบน เลือก **API Monitor Environment**
2. ตรวจสอบว่า `base_url` ถูกตั้งเป็น `http://localhost:8080`

## 📋 การใช้งาน

### ขั้นตอนที่ 1: เริ่มต้นระบบ

1. **รันเซิร์ฟเวอร์**:
   ```bash
   cd /Users/poogunkati/Desktop/WORK/api-monitor/backend
   ./api-monitor
   ```

2. **ทดสอบ Health Check**:
   - รัน request: `Health Check > Server Health Check`

### ขั้นตอนที่ 2: Authentication

1. **Login**:
   - รัน request: `Authentication > Login`
   - ใช้ username: `admin`, password: `admin123`
   - Token จะถูกบันทึกใน environment variables อัตโนมัติ

2. **Register** (ถ้าต้องการสร้าง user ใหม่):
   - รัน request: `Authentication > Register`
   - แก้ไข username, email ตามต้องการ

### ขั้นตอนที่ 3: จัดการ Endpoints

1. **ดู Endpoints ทั้งหมด**:
   - รัน request: `Endpoints Management > Get All Endpoints`

2. **สร้าง Endpoint ใหม่**:
   - รัน request: `Endpoints Management > Create Endpoint`
   - ID ของ endpoint ที่สร้างจะถูกบันทึกใน `endpoint_id`

3. **แก้ไข Endpoint**:
   - รัน request: `Endpoints Management > Update Endpoint`

4. **เปิด/ปิด Endpoint**:
   - รัน request: `Endpoints Management > Toggle Endpoint Status`

5. **ลบ Endpoint**:
   - รัน request: `Endpoints Management > Delete Endpoint`

## 🔐 Authentication

Collection นี้ใช้ **Bearer Token** authentication:

- Token จะถูกตั้งค่าอัตโนมัติหหลังจาก login สำเร็จ
- Token จะหมดอายุใน 24 ชั่วโมง
- หาก token หมดอายุ ให้ทำการ login ใหม่

## 📊 Environment Variables

| Variable | Description | Auto-Set |
|----------|-------------|----------|
| `base_url` | URL ของ API server | ❌ |
| `auth_token` | JWT token หลัง login | ✅ |
| `user_id` | ID ของ user ที่ login | ✅ |
| `username` | Username ของ user | ✅ |
| `user_role` | Role ของ user (admin/user) | ✅ |
| `endpoint_id` | ID ของ endpoint ที่สร้างล่าสุด | ✅ |

## 🧪 ตัวอย่างการทดสอบ

### การทดสอบ Flow สมบูรณ์:

1. ✅ Health Check
2. ✅ Login
3. ✅ Create Endpoint
4. ✅ Get All Endpoints
5. ✅ Update Endpoint
6. ✅ Toggle Endpoint
7. ✅ Delete Endpoint

### Sample Request Body สำหรับ Create Endpoint:

```json
{
    "name": "Test JSONPlaceholder API",
    "url": "https://jsonplaceholder.typicode.com/posts/1",
    "method": "GET",
    "headers": {
        "Content-Type": "application/json",
        "User-Agent": "API Monitor"
    },
    "body": "",
    "timeout_seconds": 30,
    "check_interval_seconds": 300,
    "is_active": true,
    "proxy_id": null
}
```

## ⚠️ หมายเหตุ

- ตรวจสอบให้แน่ใจว่าเซิร์ฟเวอร์รันอยู่บนพอร์ต 8080
- สำหรับ production ให้เปลี่ยน `base_url` เป็น URL จริง
- Token จะหมดอายุใน 24 ชั่วโมง ต้อง login ใหม่

## 🔧 Troubleshooting

**ปัญหา: 401 Unauthorized**
- ตรวจสอบว่าได้ login แล้ว
- ตรวจสอบว่า token ยังไม่หมดอายุ

**ปัญหา: 500 Internal Server Error**
- ตรวจสอบการเชื่อมต่อ database
- ตรวจสอบ JWT_SECRET ใน .env

**ปัญหา: Connection Refused**
- ตรวจสอบว่าเซิร์ฟเวอร์รันอยู่
- ตรวจสอบพอร์ต 8080 ไม่ถูกใช้โดยโปรแกรมอื่น
