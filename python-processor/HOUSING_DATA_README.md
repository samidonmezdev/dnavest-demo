# Housing Price Index Data Import

Bu script konut fiyat endeksi verilerini CSV formatından PostgreSQL veritabanına aktarmak için kullanılır.

## Özellikler

- ✅ CSV dosyalarından veri aktarımı
- ✅ Tekrarlı veri önleme (UPSERT kullanarak)
- ✅ REST API endpoint'leri
- ✅ Batch insert ile yüksek performans
- ✅ Otomatik tablo oluşturma

## Veritabanı Şeması

```sql
CREATE TABLE housing_price_index (
    id SERIAL PRIMARY KEY,
    tarih DATE NOT NULL,
    istanbul_turkiye VARCHAR(50) NOT NULL,
    yeni_yeni_olmayan_konut VARCHAR(50) NOT NULL,
    fiyat_endeksi DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tarih, istanbul_turkiye, yeni_yeni_olmayan_konut)
);
```

### Unique Constraint
`(tarih, istanbul_turkiye, yeni_yeni_olmayan_konut)` kombinasyonu unique olarak tanımlanmıştır. Bu sayede aynı tarih, lokasyon ve konut tipi için tekrarlı veri eklenmez.

## Kullanım

### 1. Standalone Script

```bash
# Örnek veri ile kullanım
python import_housing_data.py

# CSV dosyasından kullanım
python import_housing_data.py housing_data.csv
```

### 2. API Endpoint'leri

#### Veri İçe Aktarma
```bash
POST /api/housing/import

# Request body:
{
  "csv_data": "tarih,istanbul_turkiye,yeni_yeni_olmayan_konut,fiyat_endeksi\n2010-01-01,İstanbul,Yeni Konut,35.9\n..."
}

# Response:
{
  "message": "data imported successfully",
  "rows_imported": 24,
  "rows_affected": 24
}
```

#### Veri Sorgulama
```bash
GET /api/housing/data

# Query parametreleri (opsiyonel):
- location: İstanbul veya Türkiye
- type: Yeni Konut veya Yeni Olmayan Konut
- start_date: Başlangıç tarihi (YYYY-MM-DD)
- end_date: Bitiş tarihi (YYYY-MM-DD)

# Örnek:
GET /api/housing/data?location=İstanbul&type=Yeni%20Konut&start_date=2010-01-01&end_date=2010-06-01

# Response:
{
  "count": 6,
  "data": [
    {
      "id": 1,
      "tarih": "2010-01-01",
      "istanbul_turkiye": "İstanbul",
      "yeni_yeni_olmayan_konut": "Yeni Konut",
      "fiyat_endeksi": 35.9,
      "created_at": "2026-01-28T13:11:00",
      "updated_at": "2026-01-28T13:11:00"
    },
    ...
  ]
}
```

## CSV Format

CSV dosyası aşağıdaki kolonları içermelidir:

```csv
tarih,istanbul_turkiye,yeni_yeni_olmayan_konut,fiyat_endeksi
2010-01-01,İstanbul,Yeni Konut,35.9
2010-01-01,İstanbul,Yeni Olmayan Konut,35.9
2010-01-01,Türkiye,Yeni Konut,44.9
2010-01-01,Türkiye,Yeni Olmayan Konut,45.3
```

### Kolon Açıklamaları

- `tarih`: Veri tarihi (YYYY-MM-DD formatında)
- `istanbul_turkiye`: Lokasyon (İstanbul veya Türkiye)
- `yeni_yeni_olmayan_konut`: Konut tipi (Yeni Konut veya Yeni Olmayan Konut)
- `fiyat_endeksi`: Fiyat endeksi değeri (decimal)

## Tekrarlı Veri Yönetimi

Script, aynı `(tarih, istanbul_turkiye, yeni_yeni_olmayan_konut)` kombinasyonuna sahip verileri tekrar eklemez. Eğer aynı kombinasyon zaten tabloda varsa:

- Mevcut kayıt güncellenir (UPDATE)
- `fiyat_endeksi` yeni değer ile güncellenir
- `updated_at` timestamp'i güncellenir

Bu sayede aynı CSV'yi birden fazla kez çalıştırmanız durumunda tekrarlı veri oluşmaz.

## Örnek Kullanım

```python
# Python'dan kullanım
import requests

csv_data = """tarih,istanbul_turkiye,yeni_yeni_olmayan_konut,fiyat_endeksi
2010-01-01,İstanbul,Yeni Konut,35.9
2010-01-01,İstanbul,Yeni Olmayan Konut,35.9"""

response = requests.post('http://localhost:8081/api/housing/import', 
                        json={'csv_data': csv_data})
print(response.json())
```

## Environment Variables

```bash
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/microservices
```

## Dependencies

```
psycopg2-binary==2.9.9
```
