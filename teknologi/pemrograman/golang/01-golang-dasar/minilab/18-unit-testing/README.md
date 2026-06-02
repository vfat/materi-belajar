# Manual Test: Unit Testing

Folder ini berisi contoh package dan file test untuk materi **18-unit-testing.md**.

## Struktur

```
18-unit-testing/
└── mathutil/
    ├── mathutil.go
    └── mathutil_test.go
```

## Cara Menjalankan Test

Jalankan dari folder `minilab/go/18-unit-testing`:

```bash
# run all tests recursively
go test ./...

# verbose
go test -v ./...

# run a specific test
go test -run TestAdd ./...

# run benchmarks
go test ./... -bench .

# coverage
go test -cover ./...

# race detector
go test -race ./...
```

## Apa yang Ditest?

- Unit tests untuk `Add`, `Multiply`, `Divide`, `Max`, `Min`
- Table-driven tests, Example and Benchmark

