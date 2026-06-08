
# Golang + Database

> 📅 Dibuat: 2026-06-02  
> 📖 Total Materi: 20 topik

## Tentang Seri Ini

Seri ini membahas integrasi Go dengan berbagai database dan tooling pendukung: SQL (SQLite, MySQL, PostgreSQL), NoSQL (MongoDB), cache & messaging (Redis), serta integrasi GraphQL. Fokusnya praktek langsung: koneksi, query, transaksi, migrasi, dan pola integrasi ke aplikasi Go.

## Roadmap Materi

| No  | Topik                                          | File                                 | Status      | Cek Manual  |
|-----|------------------------------------------------|--------------------------------------|-------------|-------------|
| 01  | Fondasi Database di Go                         | `01-db-fundamentals.md`              | ✅ Sudah    | ✅ Lolos    |
| 02  | Golang + SQLite: Setup & CRUD                  | `02-sqlite-setup.md`                 | ✅ Sudah    | ✅ Lolos    |
| 03  | SQLite: Migrations & Schema                    | `03-sqlite-migration.md`             | ⬜ Belum    |             |
| 04  | Golang + MySQL: Connection & Config            | `04-mysql-setup.md`                  | ⬜ Belum    |             |
| 05  | MySQL: Transactions & Query Patterns           | `05-mysql-queries.md`                | ⬜ Belum    |             |
| 06  | MySQL: Migrations & Environment Sync           | `06-mysql-migration.md`              | ⬜ Belum    |             |
| 07  | Golang + PostgreSQL: Setup & JSONB             | `07-postgres-setup.md`               | ⬜ Belum    |             |
| 08  | PostgreSQL: Advanced Queries                   | `08-postgres-advanced.md`            | ⬜ Belum    |             |
| 09  | PostgreSQL: Concurrency & Locking              | `09-postgres-concurrency.md`         | ⬜ Belum    |             |
| 10  | PostgreSQL: Migrations & Backup                | `10-postgres-tooling.md`             | ⬜ Belum    |             |
| 11  | Golang + MongoDB: Setup & BSON                 | `11-mongodb-setup.md`                | ⬜ Belum    |             |
| 12  | MongoDB: CRUD & Aggregation                    | `12-mongodb-queries.md`              | ⬜ Belum    |             |
| 13  | MongoDB: Indexing & Performance                | `13-mongodb-performance.md`          | ⬜ Belum    |             |
| 14  | Golang + Redis: Setup & Basic Ops              | `14-redis-setup.md`                  | ⬜ Belum    |             |
| 15  | Redis: Caching Strategies                      | `15-redis-caching.md`                | ⬜ Belum    |             |
| 16  | Redis: Pub/Sub & Queue                         | `16-redis-pubsub.md`                 | ⬜ Belum    |             |
| 17  | GraphQL Fundamentals & gqlgen                  | `17-graphql-setup.md`                | ⬜ Belum    |             |
| 18  | GraphQL + Database Integration                 | `18-graphql-db.md`                   | ⬜ Belum    |             |
| 19  | Capstone 1: REST API dengan SQL & Redis        | `19-capstone-sql-redis.md`           | ⬜ Belum    |             |
| 20  | Capstone 2: Real-time App dengan MongoDB & GraphQL | `20-capstone-nosql-graphql.md`  | ⬜ Belum    |             |

## Catatan & Referensi

- Gunakan environment variables untuk DSN/credentials.  
- Rekomendasi driver: `database/sql` + driver khusus (sqlite3/pgx/go-sql-driver/mysql) atau driver resmi untuk NoSQL.  
- Tooling migrasi: `golang-migrate`.  

---

Mulai dari `01-db-fundamentals.md` untuk fondasi, lalu lanjut bertahap ke tiap topik.
