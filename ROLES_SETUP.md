# Rol ve Yetki Sistemi Eklendi

Auth servisine basit bir rol ve yetki sistemi eklendi.

## 🎯 Yapılan Değişiklikler

### 1. Veritabanı Şeması (`postgres/init.sql`)
- ✅ `roles` tablosu eklendi
- ✅ `user_roles` tablosu eklendi (many-to-many ilişki)
- ✅ Varsayılan roller: `admin` ve `user`
- ✅ Örnek kullanıcılara otomatik rol ataması

### 2. Model Katmanı
- ✅ `auth-service/models/role.go` - Rol yönetimi modeli
  - `GetRoleByName` - Role adına göre rol getir
  - `GetUserRoles` - Kullanıcının tüm rollerini getir
  - `AssignRoleToUser` - Kullanıcıya rol ata
  - `RemoveRoleFromUser` - Kullanıcıdan rol kaldır
  - `HasRole` - Kullanıcının belirli bir rolü olup olmadığını kontrol et

- ✅ `auth-service/models/user.go` güncellendi
  - `User` struct'ına `Roles []Role` alanı eklendi
  - `CreateUser` metodu güncellendi - yeni kullanıcılara otomatik `user` rolü atanıyor

### 3. Handler Katmanı
- ✅ `auth-service/handlers/role.go` - Rol yönetimi HTTP handler'ları
  - `GetUserRoles` - Kullanıcı rollerini listele
  - `AssignRoleToUser` - Rol ata
  - `RemoveRoleFromUser` - Rol kaldır
  - `CheckUserRole` - Rol kontrolü yap

### 4. API Endpoint'leri
```
GET  /api/roles/user?user_id=1         - Kullanıcının rollerini getir
POST /api/roles/assign                  - Kullanıcıya rol ata
POST /api/roles/remove                  - Kullanıcıdan rol kaldır
GET  /api/roles/check?user_id=1&role=admin - Kullanıcının rolünü kontrol et
```

## 📝 Kullanım Örnekleri

### Kullanıcı Rollerini Görüntüleme
```bash
curl -X GET "http://localhost:8082/api/roles/user?user_id=1"
```

### Kullanıcıya Rol Atama
```bash
curl -X POST http://localhost:8082/api/roles/assign \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "role_id": 1
  }'
```

### Kullanıcıdan Rol Kaldırma
```bash
curl -X POST http://localhost:8082/api/roles/remove \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "role_id": 2
  }'
```

### Rol Kontrolü
```bash
curl -X GET "http://localhost:8082/api/roles/check?user_id=1&role=admin"
```

## 🔄 Servisleri Yeniden Başlatma

Değişikliklerin aktif olması için servisleri yeniden başlatın:

```bash
cd "dnavest demo"
docker-compose down
docker-compose up -d --build auth-service
```

## 📊 Veritabanı Yapısı

```
users
├── id
├── email
├── password_hash
├── name
└── created_at

roles
├── id
├── name (admin, user)
├── description
└── created_at

user_roles (many-to-many)
├── user_id (FK -> users.id)
├── role_id (FK -> roles.id)
└── assigned_at
```

## ✨ Özellikler

- ✅ Basit ve anlaşılır yapı
- ✅ Many-to-many ilişki (bir kullanıcı birden fazla role sahip olabilir)
- ✅ Otomatik rol ataması (yeni kullanıcılar `user` rolüyle oluşturulur)
- ✅ Role bazlı kontrol API'si
- ✅ Cascade delete (kullanıcı silindiğinde rolleri de silinir)
