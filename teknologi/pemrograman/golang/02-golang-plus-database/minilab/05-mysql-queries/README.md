# Manual Test: MySQL Transactions & Query Patterns

Folder ini berisi script Go untuk menguji semua konsep dari materi **05-mysql-queries.md**.

## Materi yang Dicakup

- **Batch Insert** — Insert banyak data dalam 1 transaksi (`BulkInsert`)
- **Pagination** — Query dengan `LIMIT` / `OFFSET` (`ListPaginated`)
- **Aggregate Query** — `GROUP BY`, `SUM`, `AVG` (`GetCategorySummary`)
- **Row Locking** — `FOR UPDATE` untuk mencegah race condition (`CreateOrder`)
- **JOIN Query** — Gabung tabel `orders` + `customers` (`GetOrderWithCustomer`)
- **Subquery / Top N** — Aggregate dengan `LEFT JOIN` + `GROUP BY` + `ORDER BY` (`GetTopProducts`)
- **Transaksional Update** — Update status dengan validasi (`UpdateOrderStatus`)
- **Optimistic Locking** — Update dengan WHERE condition untuk deteksi konflik (`UpdateStockOptimistic`)
- **Error Handling** — Rollback otomatis saat `insufficient stock`

## Struktur File

```
05-mysql-queries/
├── docker-compose.yml   ← MySQL container
├── main.go              ← Script testing (semua test dalam 1 file)
├── README.md            ← File ini
├── go.mod               ← Module definition + dependency
└── 05-test              ← Hasil compile (setelah build, opsional)
```

## Prasyarat

- Docker & Docker Compose
- Go 1.21+
- Port 3306 tidak dipakai oleh service lain

## Cara Menjalankan

### 1. Start MySQL Container

```bash
cd 05-mysql-queries
docker compose up -d

# Tunggu sampai healthcheck lulus
docker compose ps
# Status harus "healthy"
```

### 2. Jalankan Test

```bash
cd 05-mysql-queries

# Install dependency
go mod tidy

# Jalankan langsung
go run main.go

# Atau build dulu
go build -o 05-test main.go && ./05-test
```

### 3. Cleanup

```bash
cd 05-mysql-queries
docker compose down -v
```

## Output yang Diharapkan

Program akan menjalankan 10 test secara berurutan:

1. ✅ Batch insert 8 products dalam 1 transaksi
2. ✅ Pagination page 1 & page 2 (limit=3)
3. ✅ Aggregate per category (GROUP BY)
4. ✅ Order dengan row locking (FOR UPDATE)
5. ✅ JOIN query order + customer
6. ✅ Top selling products
7. ✅ Update order status (pending → paid)
8. ✅ Optimistic lock success + detection
9. ✅ Insufficient stock error + rollback
10. ✅ Final state verification

```
=== SEMUA TEST SELESAI ===
🎉 Transactions, row locking, JOIN, pagination, aggregate, optimistic lock sudah dipahami!
```
