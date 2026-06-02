package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== MANUAL TEST MATERI 12: TIME & DATE ===\n")

	fmt.Println("--- Test 1: Now & RFC3339 ---")
	now := time.Now()
	fmt.Println("Now (RFC3339):", now.Format(time.RFC3339))

	fmt.Println("\n--- Test 2: Parse custom layout ---")
	layout := "2006-01-02 15:04:05"
	tstr := "2026-05-30 14:00:00"
	p, err := time.Parse(layout, tstr)
	if err != nil {
		fmt.Println("parse error:", err)
	} else {
		fmt.Println("Parsed:", p.Format(time.RFC3339))
	}

	fmt.Println("\n--- Test 3: Duration parsing & add/sub ---")
	d, err := time.ParseDuration("1h30m")
	if err != nil {
		fmt.Println("duration parse error:", err)
	} else {
		fmt.Println("Duration:", d.String())
		fmt.Println("Now + duration:", now.Add(d).Format("2006-01-02 15:04:05"))
		fmt.Println("(Subtract back):", now.Add(d).Add(-d).Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n--- Test 4: Timezone / Location ---")
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		fmt.Println("load location error:", err)
		loc = time.UTC
	}
	fmt.Println("Time in Asia/Jakarta:", now.In(loc).Format(time.RFC1123))

	fmt.Println("\n--- Test 5: Unix timestamps ---")
	fmt.Println("Unix(sec):", now.Unix())
	fmt.Println("Unix(nano):", now.UnixNano())

	fmt.Println("\n--- Test 6: Common layouts ---")
	fmt.Println("ANSIC:", now.Format(time.ANSIC))
	fmt.Println("RFC1123Z:", now.Format(time.RFC1123Z))

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
