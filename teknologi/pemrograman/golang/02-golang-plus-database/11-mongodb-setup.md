---
topik: Golang + MongoDB Setup & BSON
urutan: 11 dari 20
posisi: lanjutan
sebelumnya: PostgreSQL Migrations & Backup
prerequisites:
  - PostgreSQL Migrations & Backup (10-postgres-tooling)
  - Golang + PostgreSQL Setup & JSONB (07-postgres-setup)
level: menengah
---

> 🔗 **Lanjutan dari:** PostgreSQL Migrations & Backup
> ← Kembali ke: `10-postgres-tooling.md`

# Golang + MongoDB: Setup & BSON

## Tujuan Pembelajaran

Setelah materi ini, kamu akan:

- Menjalankan MongoDB via Docker Compose
- Menghubungkan Go ke MongoDB menggunakan driver resmi `mongo-driver`
- Memahami perbedaan konsep **document database** vs relational database
- Menggunakan **BSON** tags untuk mapping struct Go ke document MongoDB
- Melakukan CRUD dasar pada collection MongoDB
- Menangani tipe khas MongoDB: `ObjectID`, embedded document, array, timestamp
- Mengatur connection timeout, ping, dan graceful shutdown
- Menyusun struktur repository sederhana untuk aplikasi Go + MongoDB

---

## 1. Mengapa MongoDB?

MongoDB adalah **NoSQL document database** yang menyimpan data dalam bentuk **document BSON**. Cocok untuk data yang bentuknya fleksibel, nested, dan cepat berubah.

| Kebutuhan | PostgreSQL | MongoDB |
|---|---|---|
| Schema ketat & relasi kuat | ✅ Sangat cocok | ⚠️ Kurang ideal |
| Document nested | ⚠️ Bisa via JSONB | ✅ Native |
| Data fleksibel / evolving schema | ⚠️ Butuh migration | ✅ Sangat cocok |
| Join kompleks | ✅ Kuat | ⚠️ Terbatas (`$lookup`) |
| Transaksi multi-record | ✅ Mature | ✅ Ada, tapi bukan default utama |
| Horizontal scaling | ✅ Bisa | ✅ Umum dipakai |

> 💡 **Pilih MongoDB** saat struktur data sering berubah, banyak nested object/array, atau saat aplikasi lebih cocok memakai model document daripada tabel relasional.

---

## 2. Setup Docker Compose

```yaml
# docker-compose.yml
services:
  mongodb:
    image: mongo:7
    container_name: golang_mongodb_demo
    restart: unless-stopped
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: root
      MONGO_INITDB_DATABASE: golang_demo
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--username", "root", "--password", "root", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mongodb_data:
    driver: local
```

```bash
# Jalankan
docker compose up -d

# Cek status (tunggu healthy)
docker compose ps

# Masuk ke mongosh (opsional)
docker exec -it golang_mongodb_demo mongosh -u root -p root
```

---

## 3. Driver Resmi MongoDB untuk Go

MongoDB menyediakan driver resmi:

| Driver | Import | Status |
|---|---|---|
| Official MongoDB Driver | `go.mongodb.org/mongo-driver/mongo` | ✅ Direkomendasikan |

```bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
go get go.mongodb.org/mongo-driver/bson/primitive
```

> 💡 Driver ini bekerja dengan `context.Context`, mendukung timeout, session, transaction, dan API BSON yang lengkap.

---

## 4. Koneksi & URI MongoDB

### 4.1 Format Connection URI

```text
mongodb://username:password@host:port/database?authSource=admin
```

Contoh untuk local Docker:

```text
mongodb://root:root@localhost:27017/golang_demo?authSource=admin
```

| Bagian | Contoh | Keterangan |
|---|---|---|
| username | `root` | User MongoDB |
| password | `root` | Password MongoDB |
| host | `localhost` | Host database |
| port | `27017` | Port default MongoDB |
| database | `golang_demo` | Database yang dipakai |
| `authSource=admin` | `admin` | DB untuk autentikasi root user |

### 4.2 Koneksi Dasar di Go

```go
package db

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
    URI      string
    DBName   string
    Timeout  time.Duration
}

func Connect(ctx context.Context, cfg Config) (*mongo.Client, error) {
    clientOpts := options.Client().ApplyURI(cfg.URI)

    connectCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
    defer cancel()

    client, err := mongo.Connect(connectCtx, clientOpts)
    if err != nil {
        return nil, fmt.Errorf("connect mongodb: %w", err)
    }

    pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
    defer pingCancel()

    if err := client.Ping(pingCtx, nil); err != nil {
        _ = client.Disconnect(ctx)
        return nil, fmt.Errorf("ping mongodb: %w", err)
    }

    return client, nil
}
```

---

## 5. Konsep BSON dan Struct Tags

MongoDB menyimpan data sebagai **BSON** (*Binary JSON*). Di Go, field struct biasanya diberi tag `bson`.

### 5.1 Contoh Struct Dasar

```go
package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Name      string             `bson:"name" json:"name"`
    Email     string             `bson:"email" json:"email"`
    Age       int                `bson:"age" json:"age"`
    IsActive  bool               `bson:"is_active" json:"is_active"`
    Tags      []string           `bson:"tags,omitempty" json:"tags,omitempty"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
```

### 5.2 Arti Tag BSON Penting

| Tag | Arti |
|---|---|
| `bson:"_id"` | Mapping ke primary key document MongoDB |
| `omitempty` | Jangan simpan field jika zero-value |
| `bson:"name"` | Nama field di document |

> 💡 Jika field `_id` bertipe `primitive.ObjectID` dan nilainya kosong, MongoDB akan membuatkannya otomatis saat insert.

---

## 6. Tipe Khas MongoDB

### 6.1 `ObjectID`

MongoDB default primary key adalah `ObjectID`.

```go
id := primitive.NewObjectID()
fmt.Println(id.Hex())

parsedID, err := primitive.ObjectIDFromHex("6652f497f7b2b0b0c4d4e8af")
if err != nil {
    return err
}
```

### 6.2 Embedded Document

```go
type Address struct {
    Street  string `bson:"street"`
    City    string `bson:"city"`
    Country string `bson:"country"`
}

type Customer struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    Name    string             `bson:"name"`
    Address Address            `bson:"address"`
}
```

### 6.3 Array

```go
type Product struct {
    ID         primitive.ObjectID `bson:"_id,omitempty"`
    Name       string             `bson:"name"`
    Categories []string           `bson:"categories"`
    Images     []string           `bson:"images,omitempty"`
}
```

### 6.4 Flexible Field dengan `bson.M`

```go
metadata := bson.M{
    "source": "importer",
    "score":  98,
    "flags":  []string{"featured", "verified"},
}
```

---

## 7. CRUD Dasar

### 7.1 Insert One

```go
user := User{
    Name:      "Alice",
    Email:     "alice@example.com",
    Age:       28,
    IsActive:  true,
    Tags:      []string{"golang", "mongodb"},
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

result, err := usersColl.InsertOne(ctx, user)
if err != nil {
    return err
}

fmt.Println("Inserted ID:", result.InsertedID)
```

### 7.2 Find One

```go
var user User
err := usersColl.FindOne(ctx, bson.M{"email": "alice@example.com"}).Decode(&user)
if err != nil {
    if errors.Is(err, mongo.ErrNoDocuments) {
        return fmt.Errorf("user tidak ditemukan")
    }
    return err
}
```

### 7.3 Find Many

```go
cursor, err := usersColl.Find(ctx, bson.M{"is_active": true})
if err != nil {
    return err
}
defer cursor.Close(ctx)

var users []User
if err := cursor.All(ctx, &users); err != nil {
    return err
}
```

### 7.4 Update One

```go
filter := bson.M{"email": "alice@example.com"}
update := bson.M{
    "$set": bson.M{
        "age":        29,
        "updated_at": time.Now(),
    },
    "$push": bson.M{
        "tags": "backend",
    },
}

result, err := usersColl.UpdateOne(ctx, filter, update)
if err != nil {
    return err
}

fmt.Println("Matched:", result.MatchedCount)
fmt.Println("Modified:", result.ModifiedCount)
```

### 7.5 Delete One

```go
result, err := usersColl.DeleteOne(ctx, bson.M{"email": "alice@example.com"})
if err != nil {
    return err
}

fmt.Println("Deleted:", result.DeletedCount)
```

---

## 8. Repository Pattern Sederhana

```go
package repository

import (
    "context"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Name      string             `bson:"name"`
    Email     string             `bson:"email"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}

type UserRepository struct {
    coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
    return &UserRepository{coll: db.Collection("users")}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    now := time.Now()
    user.CreatedAt = now
    user.UpdatedAt = now

    res, err := r.coll.InsertOne(ctx, user)
    if err != nil {
        return fmt.Errorf("insert user: %w", err)
    }

    if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
        user.ID = oid
    }
    return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
        }
        return nil, fmt.Errorf("find user by email: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) UpdateName(ctx context.Context, id primitive.ObjectID, name string) error {
    _, err := r.coll.UpdateOne(ctx,
        bson.M{"_id": id},
        bson.M{"$set": bson.M{"name": name, "updated_at": time.Now()}},
    )
    if err != nil {
        return fmt.Errorf("update user: %w", err)
    }
    return nil
}

func (r *UserRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
    _, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return fmt.Errorf("delete user: %w", err)
    }
    return nil
}
```

---

## 9. Index Dasar

MongoDB tidak punya schema ketat seperti SQL, tapi **index** tetap sangat penting.

```go
import "go.mongodb.org/mongo-driver/mongo/options"

indexModel := mongo.IndexModel{
    Keys: bson.D{{Key: "email", Value: 1}},
    Options: options.Index().SetUnique(true).SetName("uniq_users_email"),
}

name, err := usersColl.Indexes().CreateOne(ctx, indexModel)
if err != nil {
    return err
}

fmt.Println("index created:", name)
```

Contoh index umum:

| Index | Kegunaan |
|---|---|
| `{ email: 1 }` | Lookup berdasarkan email |
| `{ created_at: -1 }` | Sorting data terbaru |
| `{ status: 1, created_at: -1 }` | Filter + sort |
| Unique index | Cegah duplikasi |

---

## 10. Timeout, Context, dan Graceful Shutdown

Karena driver MongoDB sangat bergantung pada `context.Context`, selalu beri timeout untuk operasi penting.

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
if err != nil {
    return err
}

defer func() {
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()
    _ = client.Disconnect(shutdownCtx)
}()
```

> ⚠️ Tanpa timeout, operasi ke database bisa menggantung saat jaringan bermasalah atau server overload.

---

## 11. Perbedaan Mindset SQL vs MongoDB

| SQL / PostgreSQL | MongoDB |
|---|---|
| Table | Collection |
| Row | Document |
| Column | Field |
| Join | Embedded document / reference / `$lookup` |
| Schema ketat | Schema fleksibel |
| Foreign key | Dikelola aplikasi |
| `SERIAL` / `UUID` | `ObjectID` |

Contoh pendekatan:

- **SQL mindset:** pisahkan `users`, `addresses`, `phones` ke tabel berbeda.
- **MongoDB mindset:** simpan `address` dan `phones` langsung di dalam document user jika memang selalu dibaca bersama.

---

## 12. Best Practices

- Gunakan **driver resmi MongoDB**
- Simpan `_id` sebagai `primitive.ObjectID`
- Selalu beri **timeout** saat connect/query
- Buat **index** untuk field yang sering dicari
- Jangan terlalu banyak nested document jika sering diupdate sebagian
- Gunakan embedded document untuk data yang selalu dibaca bersama
- Gunakan reference jika document tumbuh terlalu besar atau relasi sangat banyak
- Validasi schema di level aplikasi walau MongoDB fleksibel

---

## 13. Ringkasan

| Topik | Kunci Utama |
|---|---|
| Driver | `mongo-driver` resmi |
| Connection URI | `mongodb://user:pass@host:27017/db?authSource=admin` |
| Primary Key | `primitive.ObjectID` |
| Mapping Struct | Tag `bson` |
| Query Dinamis | `bson.M`, `bson.D` |
| CRUD | `InsertOne`, `FindOne`, `Find`, `UpdateOne`, `DeleteOne` |
| Nested Data | Embedded document + array |
| Stabilitas | `context.WithTimeout`, `Ping`, `Disconnect` |
| Performance | Index di field penting |

---

## 14. Latihan

1. Buat `Product` struct dengan field `_id`, `name`, `price`, `stock`, `tags`
2. Insert 5 document product ke collection `products`
3. Buat query untuk mencari product berdasarkan tag tertentu
4. Tambahkan unique index pada field `email` di collection `users`
5. Buat repository `ProductRepository` dengan method `Create`, `FindAll`, `FindByID`, `UpdateStock`, `Delete`
6. Simpan embedded document `address` di collection `customers`

```bash
# Jalankan MongoDB
docker compose up -d

# Install dependency
go get go.mongodb.org/mongo-driver/mongo

# Jalankan aplikasi
go run main.go
```

---

## 15. Referensi

- https://www.mongodb.com/docs/manual/introduction/ — MongoDB Introduction
- https://www.mongodb.com/docs/drivers/go/current/ — MongoDB Go Driver
- https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo — Package `mongo`
- https://pkg.go.dev/go.mongodb.org/mongo-driver/bson — Package `bson`
- https://www.mongodb.com/docs/manual/core/document/ — BSON Document Structure

---

> ⏭️ **Selanjutnya:** `12-mongodb-queries.md` — MongoDB: CRUD & Aggregation
