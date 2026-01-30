# Microservices Architecture

A complete Go-based microservices architecture with API Gateway, Authentication Service, Go API Service, Python Data Processor, PostgreSQL, and Redis.

## 🏗️ Architecture

```
┌─────────────────┐
│   API Gateway   │ :8000
│   (Go)          │
└────────┬────────┘
         │
    ┌────┴────┬─────────────┬──────────────┐
    │         │             │              │
┌───▼────┐ ┌──▼─────┐ ┌────▼─────┐ ┌─────▼──────┐
│  Auth  │ │ Go API │ │  Python  │ │   Redis    │
│Service │ │:8080   │ │Processor │ │   :6379    │
│ :8082  │ │        │ │  :8081   │ │            │
└───┬────┘ └───┬────┘ └────┬─────┘ └─────┬──────┘
    │          │           │              │
    └──────────┴───────────┴──────────────┘
                     │
              ┌──────▼───────┐
              │  PostgreSQL  │
              │    :5432     │
              └──────────────┘
```



🔹 1. Basit Konut Fiyat Endeksi Paneli (Önerilen Ana Proje)
🎯 Proje Amacı

“İstanbul ve Türkiye’de yeni ve eski konutların fiyat endeksi zaman içinde nasıl değişmiş?”

📊 Panelde Olabilecek Bölümler
1️⃣ Filtreler (çok önemli)

📍 Bölge: İstanbul / Türkiye

🏠 Konut Türü: Yeni / Yeni Olmayan

📅 Tarih Aralığı (slider ya da dropdown)

2️⃣ Temel Göstergeler (KPI)

Son ay fiyat endeksi

İlk yıla göre % değişim

Son 1 yıldaki artış oranı

En yüksek / en düşük değer

3️⃣ Grafikler

📈 Zaman Serisi Grafiği

Tarihe göre fiyat endeksi

📊 Karşılaştırma Grafiği

İstanbul vs Türkiye

🏘 Yeni vs Yeni Olmayan Konut

Aynı grafikte iki çizgi

🔹 2. Yapımı Çok Kolay Analiz Fikirleri
✔ Aylık ve Yıllık Değişim Analizi

Ay bazında % artış

Yıllık ortalama fiyat endeksi

✔ İstanbul – Türkiye Farkı

Aynı tarihte İstanbul ile Türkiye arasındaki fark

Bu fark zamanla açılıyor mu kapanıyor mu?

✔ Yeni Konutlar Daha mı Hızlı Artıyor?

Yeni konut endeksi vs yeni olmayan

Hangisi daha volatil?

🔹 3. Kullanabileceğin Teknolojiler (Basitten Zora)
🟢 Çok Basit (Öğrenci dostu)

Excel / Google Sheets

Pivot table + grafik = mini panel

🟡 Orta Seviye (çok iyi CV katkısı)

Python + Streamlit

pandas + matplotlib / plotly

CV’de çok iyi durur:
“Konut Fiyat Endeksi Dashboard’u geliştirdim”

🔵 Alternatif

Power BI

Tableau Public

🔹 4. Mini Proje Tanımı (Birebir Kullanmalık)

Proje Adı: Türkiye ve İstanbul Konut Fiyat Endeksi Analizi
Açıklama:
Bu projede 2010 sonrası dönemde Türkiye ve İstanbul’da yeni ve yeni olmayan konutların fiyat endeksleri analiz edilmiştir. Kullanıcılar tarih, bölge ve konut türüne göre filtreleme yaparak fiyat değişimlerini görselleştirebilmektedir.


## 🚀 Features

### API Gateway (Go)
- ⚡ **Rate Limiting**: Token bucket algorithm with Redis
- 🛡️ **Circuit Breaker**: Prevents cascading failures
- 🔐 **JWT Authentication**: Validates tokens for protected routes
- 🔄 **Load Balancing**: Round-robin distribution
- 📝 **Request Logging**: Comprehensive request/response logs

### Auth Service (Go)
- 🔑 JWT-based authentication
- 🔒 Bcrypt password hashing (cost: 12)
- ♻️ Refresh token support
- 💾 Redis session management
- ✅ Input validation

### Go API Service
- 📊 PostgreSQL integration
- ⚡ Redis caching
- 📈 Statistics endpoints
- 🔐 Protected routes with JWT

### Python Data Processor
- 🎯 Async job processing
- 📋 Job queue with Redis
- 💾 Result persistence in PostgreSQL
- 📊 Processing statistics

## 📋 Prerequisites

- Docker & Docker Compose
- Git

## 🛠️ Installation

1. **Clone or navigate to the project directory:**
```bash
cd "dnavest demo"
```

2. **Start all services:**
```bash
docker-compose up -d
```

3. **Check service health:**
```bash
docker-compose ps
```

4. **View logs:**
```bash
docker-compose logs -f
```

## 📚 API Documentation

### Authentication Endpoints

#### Register
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!",
    "name": "John Doe"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePass123!"
  }'
```

Response:
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
  "expiresIn": 900
}
```

#### Refresh Token
```bash
curl -X POST http://localhost:8000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "your-refresh-token"
  }'
```

#### Verify Token
```bash
curl -X GET http://localhost:8000/api/v1/auth/verify \
  -H "Authorization: Bearer your-access-token"
```

### Go API Endpoints

#### Get Cached Data (Protected)
```bash
TOKEN="your-access-token"
curl -X GET http://localhost:8000/api/v1/data \
  -H "Authorization: Bearer $TOKEN"
```

#### Get Statistics (Protected)
```bash
curl -X GET http://localhost:8000/api/v1/data/stats \
  -H "Authorization: Bearer $TOKEN"
```

### Python Processor Endpoints

#### Submit Processing Job
```bash
curl -X POST http://localhost:8000/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "data": "Your data to process"
  }'
```

Response:
```json
{
  "message": "job queued successfully",
  "job_id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "queued"
}
```

#### Check Job Status
```bash
curl -X GET http://localhost:8000/api/v1/process/jobs/{job_id}
```

## 🔧 Configuration

Edit `.env` file to customize:

- **Database credentials**
- **JWT secret** (IMPORTANT: Change in production!)
- **Service ports**
- **Token expiration times**

## 🧪 Testing

### Complete Test Flow
```bash
# 1. Register a user
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!","name":"Test User"}'

# 2. Login and capture token
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!"}' | jq -r '.accessToken')

echo "Token: $TOKEN"

# 3. Access protected endpoint
curl -X GET http://localhost:8000/api/v1/data \
  -H "Authorization: Bearer $TOKEN"

# 4. Submit processing job
JOB_RESPONSE=$(curl -s -X POST http://localhost:8000/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{"data":"Hello World"}')

echo "$JOB_RESPONSE"

# 5. Get job ID and check status
JOB_ID=$(echo "$JOB_RESPONSE" | jq -r '.job_id')
sleep 4
curl -X GET "http://localhost:8000/api/v1/process/jobs/$JOB_ID"
```

## 📊 Monitoring

### Service Health Checks
```bash
# API Gateway
curl http://localhost:8000/health

# Auth Service
curl http://localhost:8082/health

# Go API
curl http://localhost:8080/health

# Python Processor
curl http://localhost:8081/health
```

### Database Access
```bash
docker exec -it microservices-postgres psql -U postgres -d microservices
```

### Redis Access
```bash
docker exec -it microservices-redis redis-cli
```

## 🛑 Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: Deletes all data)
docker-compose down -v
```

## 🔍 Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose logs service-name

# Rebuild service
docker-compose up -d --build service-name
```

### Database connection issues
```bash
# Check if PostgreSQL is ready
docker-compose logs postgres

# Reset database
docker-compose down -v
docker-compose up -d
```

### Port conflicts
Edit `.env` file and change the conflicting port numbers.

## 🏗️ Development

### Rebuild specific service
```bash
docker-compose up -d --build auth-service
```

### View real-time logs
```bash
docker-compose logs -f api-gateway
```

## 📝 Project Structure

```
.
├── api-gateway/          # Go API Gateway
│   ├── main.go
│   ├── middleware/
│   ├── router/
│   ├── go.mod
│   └── Dockerfile
├── auth-service/        # Go Auth Service
│   ├── main.go
│   ├── handlers/
│   ├── models/
│   ├── go.mod
│   └── Dockerfile
├── go-api/             # Go API Service
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── python-processor/   # Python Data Processor
│   ├── processor.py
│   ├── requirements.txt
│   └── Dockerfile
├── postgres/          # Database initialization
│   └── init.sql
├── docker-compose.yml
├── .env
└── README.md
```

## 🔐 Security Notes

- **Change JWT_SECRET** in production
- Use **strong passwords** for database
- Enable **HTTPS/TLS** in production
- Implement **rate limiting** per user
- Use **environment-specific** configurations

## 📄 License

MIT

## 👥 Contributors

Your Name
