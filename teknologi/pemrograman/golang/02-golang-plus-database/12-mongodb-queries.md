---
topik: MongoDB CRUD & Aggregation
urutan: 12 dari 20
posisi: lanjutan
sebelumnya: Golang + MongoDB Setup & BSON
prerequisites:
  - Golang + MongoDB Setup & BSON (11-mongodb-setup)
level: menengah-lanjut
---

> 🔗 **Lanjutan dari:** Golang + MongoDB Setup & BSON
> ← Kembali ke: `11-mongodb-setup.md`

# MongoDB: CRUD & Aggregation

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Melakukan **CRUD lanjutan** di MongoDB dengan Go
- Menggunakan filter query yang lebih kaya: comparison, logical, array, dan nested field
- Memakai **projection**, **sorting**, **pagination**, dan **bulk write**
- Mengimplementasikan **upsert** dan update operator MongoDB
- Menulis **aggregation pipeline** dengan `$match`, `$group`, `$sort`, `$project`, `$unwind`, dan `$lookup`
- Memahami kapan memakai query biasa vs aggregation
- Membangun repository method yang lebih fleksibel untuk query dinamis

---

## 1. Recap: CRUD Dasar vs CRUD Lanjutan

Di materi sebelumnya, kita sudah memakai `InsertOne`, `FindOne`, `Find`, `UpdateOne`, dan `DeleteOne`.

Sekarang kita lanjut ke pola yang lebih sering dipakai di aplikasi nyata:

| Kasus | Operasi MongoDB |
|---|---|
| Cari user aktif umur > 21 | `Find` + filter comparison |
| Ambil 10 data terbaru | `Find` + `Sort` + `Limit` |
| Update atau buat jika belum ada | `UpdateOne` + `Upsert` |
| Update banyak document sekaligus | `UpdateMany` |
| Insert/update/delete dalam batch | `BulkWrite` |
| Rekap total penjualan per kategori | `Aggregate` + `$group` |
| Join orders ke users | `Aggregate` + `$lookup` |

---

## 2. Filter Query yang Lebih Kaya

MongoDB query filter umumnya ditulis dengan `bson.M` atau `bson.D`.

### 2.1 Comparison Operators

```go
filter := bson.M{
    "age": bson.M{
        "$gte": 21,
        "$lte": 35,
    },
    "is_active": true,
}
```

Operator umum:

| Operator | Arti |
|---|---|
| `$eq` | sama dengan |
| `$ne` | tidak sama dengan |
| `$gt` / `$gte` | lebih besar / lebih besar sama dengan |
| `$lt` / `$lte` | lebih kecil / lebih kecil sama dengan |
| `$in` | nilai termasuk dalam array |
| `$nin` | nilai tidak termasuk dalam array |

### 2.2 Logical Operators

```go
filter := bson.M{
    "$or": []bson.M{
        {"role": "admin"},
        {"is_active": true},
    },
}
```

```go
filter := bson.M{
    "$and": []bson.M{
        {"age": bson.M{"$gte": 18}},
        {"email_verified": true},
    },
}
```

### 2.3 Query Array

```go
filter := bson.M{
    "tags": bson.M{"$in": []string{"golang", "backend"}},
}
```

Jika ingin semua tag ada:

```go
filter := bson.M{
    "tags": bson.M{"$all": []string{"golang", "mongodb"}},
}
```

### 2.4 Query Nested Field

```go
filter := bson.M{
    "address.city": "Jakarta",
}
```

---

## 3. Projection, Sort, Limit, Skip

### 3.1 Projection

Projection dipakai untuk memilih field tertentu.

```go
opts := options.Find().SetProjection(bson.M{
    "name":  1,
    "email": 1,
})

cursor, err := usersColl.Find(ctx, bson.M{}, opts)
```

> 💡 Secara default `_id` tetap ikut, kecuali di-set `0`.

```go
opts := options.Find().SetProjection(bson.M{
    "_id":   0,
    "name":  1,
    "email": 1,
})
```

### 3.2 Sorting

```go
opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
```

| Nilai | Arti |
|---|---|
| `1` | ascending |
| `-1` | descending |

### 3.3 Pagination

```go
page := int64(2)
limit := int64(10)
skip := (page - 1) * limit

opts := options.Find().
    SetSort(bson.D{{Key: "created_at", Value: -1}}).
    SetSkip(skip).
    SetLimit(limit)

cursor, err := usersColl.Find(ctx, bson.M{"is_active": true}, opts)
```

---

## 4. Update Operators yang Sering Dipakai

MongoDB update memakai operator seperti `$set`, `$inc`, `$push`, `$pull`, dan `$unset`.

### 4.1 `$set`

```go
update := bson.M{
    "$set": bson.M{
        "name":       "Alice Updated",
        "updated_at": time.Now(),
    },
}
```

### 4.2 `$inc`

```go
update := bson.M{
    "$inc": bson.M{
        "stock": -1,
        "views": 1,
    },
}
```

### 4.3 `$push` dan `$each`

```go
update := bson.M{
    "$push": bson.M{
        "tags": bson.M{
            "$each": []string{"api", "microservice"},
        },
    },
}
```

### 4.4 `$pull`

```go
update := bson.M{
    "$pull": bson.M{
        "tags": "deprecated",
    },
}
```

### 4.5 `$unset`

```go
update := bson.M{
    "$unset": bson.M{
        "legacy_field": "",
    },
}
```

---

## 5. Upsert

**Upsert** berarti: kalau document belum ada, buat baru; kalau ada, update.

```go
filter := bson.M{"email": "alice@example.com"}
update := bson.M{
    "$set": bson.M{
        "name":       "Alice",
        "updated_at": time.Now(),
    },
    "$setOnInsert": bson.M{
        "created_at": time.Now(),
        "is_active":  true,
    },
}

opts := options.Update().SetUpsert(true)
result, err := usersColl.UpdateOne(ctx, filter, update, opts)
if err != nil {
    return err
}

fmt.Println("Matched:", result.MatchedCount)
fmt.Println("Modified:", result.ModifiedCount)
fmt.Println("UpsertedID:", result.UpsertedID)
```

---

## 6. Bulk Write

Untuk banyak operasi sekaligus, pakai `BulkWrite`.

```go
models := []mongo.WriteModel{
    mongo.NewInsertOneModel().SetDocument(bson.M{
        "name":  "Laptop Pro",
        "price": 15000000,
        "stock": 10,
    }),
    mongo.NewUpdateOneModel().
        SetFilter(bson.M{"slug": "kopi-arabika"}).
        SetUpdate(bson.M{"$inc": bson.M{"stock": 20}}),
    mongo.NewDeleteOneModel().
        SetFilter(bson.M{"slug": "produk-lama"}),
}

res, err := productsColl.BulkWrite(ctx, models)
if err != nil {
    return err
}

fmt.Println("Inserted:", res.InsertedCount)
fmt.Println("Modified:", res.ModifiedCount)
fmt.Println("Deleted:", res.DeletedCount)
```

---

## 7. Aggregation Pipeline

Aggregation adalah fitur penting MongoDB untuk transformasi dan analisis data.

Pipeline disusun bertahap, misalnya:
- `$match` → filter data
- `$group` → agregasi
- `$sort` → urutkan
- `$project` → pilih/bentuk field
- `$unwind` → pecah array jadi banyak row
- `$lookup` → join ke collection lain

### 7.1 Aggregation Dasar: Rekap per Kategori

Misal collection `products` punya field:

```json
{
  "name": "Laptop Pro",
  "category": "Elektronik",
  "price": 15000000,
  "stock": 10,
  "tags": ["gadget", "premium"]
}
```

Pipeline:

```go
pipeline := mongo.Pipeline{
    {{Key: "$group", Value: bson.D{
        {Key: "_id", Value: "$category"},
        {Key: "total_products", Value: bson.D{{Key: "$sum", Value: 1}}},
        {Key: "avg_price", Value: bson.D{{Key: "$avg", Value: "$price"}}},
        {Key: "total_stock", Value: bson.D{{Key: "$sum", Value: "$stock"}}},
    }}},
    {{Key: "$sort", Value: bson.D{{Key: "avg_price", Value: -1}}}},
}

cursor, err := productsColl.Aggregate(ctx, pipeline)
if err != nil {
    return err
}
```

Hasil decode bisa ke struct:

```go
type CategorySummary struct {
    Category      string  `bson:"_id"`
    TotalProducts int32   `bson:"total_products"`
    AvgPrice      float64 `bson:"avg_price"`
    TotalStock    int32   `bson:"total_stock"`
}
```

### 7.2 `$match` + `$group`

```go
pipeline := mongo.Pipeline{
    {{Key: "$match", Value: bson.D{{Key: "is_active", Value: true}}}},
    {{Key: "$group", Value: bson.D{
        {Key: "_id", Value: "$role"},
        {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
    }}},
}
```

### 7.3 `$unwind` untuk Array

Misal setiap product punya array `tags`.

```go
pipeline := mongo.Pipeline{
    {{Key: "$unwind", Value: "$tags"}},
    {{Key: "$group", Value: bson.D{
        {Key: "_id", Value: "$tags"},
        {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
    }}},
    {{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
}
```

### 7.4 `$project`

```go
pipeline := mongo.Pipeline{
    {{Key: "$project", Value: bson.D{
        {Key: "name", Value: 1},
        {Key: "price", Value: 1},
        {Key: "price_with_tax", Value: bson.D{{Key: "$multiply", Value: bson.A{"$price", 1.11}}}},
    }}},
}
```

### 7.5 `$lookup` (Join Sederhana)

Misal collection `orders` menyimpan `user_id`, dan ingin join ke `users`.

```go
pipeline := mongo.Pipeline{
    {{Key: "$lookup", Value: bson.D{
        {Key: "from", Value: "users"},
        {Key: "localField", Value: "user_id"},
        {Key: "foreignField", Value: "_id"},
        {Key: "as", Value: "user"},
    }}},
    {{Key: "$unwind", Value: "$user"}},
    {{Key: "$project", Value: bson.D{
        {Key: "total", Value: 1},
        {Key: "status", Value: 1},
        {Key: "user_name", Value: "$user.name"},
        {Key: "user_email", Value: "$user.email"},
    }}},
}
```

> ⚠️ `$lookup` berguna, tapi jika terlalu sering dipakai untuk relasi kompleks, evaluasi lagi apakah model document sudah tepat.

---

## 8. Kapan Pakai Query Biasa vs Aggregation?

| Kebutuhan | Gunakan |
|---|---|
| Ambil data langsung berdasarkan filter sederhana | `Find` / `FindOne` |
| Update/delete langsung | `UpdateOne`, `DeleteOne`, dst |
| Rekap, statistik, grouping | `Aggregate` |
| Transformasi field / reshape output | `Aggregate` + `$project` |
| Join antar collection | `Aggregate` + `$lookup` |

---

## 9. Repository Pattern untuk Query Dinamis

```go
type ProductFilter struct {
    Category string
    MinPrice *float64
    MaxPrice *float64
    Tags     []string
    ActiveOnly bool
    Limit    int64
    Offset   int64
}

func (r *ProductRepository) FindAll(ctx context.Context, f ProductFilter) ([]Product, error) {
    filter := bson.M{}

    if f.Category != "" {
        filter["category"] = f.Category
    }

    if f.MinPrice != nil || f.MaxPrice != nil {
        priceFilter := bson.M{}
        if f.MinPrice != nil {
            priceFilter["$gte"] = *f.MinPrice
        }
        if f.MaxPrice != nil {
            priceFilter["$lte"] = *f.MaxPrice
        }
        filter["price"] = priceFilter
    }

    if len(f.Tags) > 0 {
        filter["tags"] = bson.M{"$in": f.Tags}
    }

    if f.ActiveOnly {
        filter["is_active"] = true
    }

    opts := options.Find().
        SetSort(bson.D{{Key: "created_at", Value: -1}}).
        SetLimit(f.Limit).
        SetSkip(f.Offset)

    cursor, err := r.coll.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var products []Product
    if err := cursor.All(ctx, &products); err != nil {
        return nil, err
    }
    return products, nil
}
```

---

## 10. Best Practices

- Gunakan `bson.D` jika urutan field penting, terutama di aggregation stage tertentu
- Gunakan `bson.M` untuk filter sederhana yang tidak butuh urutan khusus
- Selalu batasi hasil query dengan `Limit` jika data bisa besar
- Pastikan field yang sering dipakai di filter/sort punya index
- Hati-hati dengan `Skip` besar karena bisa mahal; untuk dataset besar pertimbangkan cursor-based pagination
- Gunakan `$project` untuk mengurangi data yang tidak perlu dikirim
- Jangan terlalu banyak memakai `$lookup` jika model document bisa disederhanakan

---

## 11. Ringkasan

| Topik | Kunci Utama |
|---|---|
| Filter Query | `$gte`, `$lte`, `$in`, `$or`, nested field |
| Projection | Pilih field tertentu |
| Sorting & Paging | `SetSort`, `SetSkip`, `SetLimit` |
| Update Operators | `$set`, `$inc`, `$push`, `$pull`, `$unset` |
| Upsert | `SetUpsert(true)` |
| Bulk Write | Banyak operasi sekaligus |
| Aggregation | `$match`, `$group`, `$sort`, `$project`, `$unwind`, `$lookup` |
| Repository Dinamis | Build filter dari input struct |

---

## 12. Latihan

1. Buat query `Find` untuk product dengan `price >= 100000` dan `stock > 0`
2. Tambahkan pagination untuk list users aktif
3. Buat update yang menambahkan tag baru ke product tanpa menghapus tag lama
4. Buat upsert untuk user berdasarkan email
5. Buat aggregation untuk menghitung jumlah product per category
6. Buat aggregation untuk menghitung tag yang paling sering muncul
7. Buat `$lookup` sederhana antara `orders` dan `users`

```bash
# Jalankan MongoDB
docker compose up -d

# Install dependency
go get go.mongodb.org/mongo-driver/mongo

# Jalankan aplikasi
go run main.go
```

---

## 13. Referensi

- https://www.mongodb.com/docs/manual/crud/ — MongoDB CRUD Operations
- https://www.mongodb.com/docs/manual/aggregation/ — Aggregation Pipeline
- https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/ — MongoDB Go Driver CRUD
- https://www.mongodb.com/docs/drivers/go/current/fundamentals/aggregation/ — MongoDB Go Driver Aggregation
- https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo/options — Package `options`

---

> ⏭️ **Selanjutnya:** `13-mongodb-performance.md` — MongoDB: Indexing & Performance
