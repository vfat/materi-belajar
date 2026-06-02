package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		fmt.Printf("worker %d processing job %d\n", id, j)
		time.Sleep(50 * time.Millisecond)
		results <- j * 2
	}
}

func main() {
	fmt.Println("=== MANUAL TEST MATERI 16: GOROUTINE & CHANNEL ===\n")

	fmt.Println("--- Test 1: Simple goroutine and done channel ---")
	done := make(chan struct{})
	go func() {
		fmt.Println("hello from goroutine")
		close(done)
	}()
	<-done

	fmt.Println("\n--- Test 2: Unbuffered & buffered channels ---")
	ch := make(chan int)
	go func() { ch <- 1 }()
	v := <-ch
	fmt.Println("received unbuffered:", v)

	bch := make(chan int, 2)
	bch <- 10
	bch <- 20
	fmt.Println("buffered received:", <-bch, <-bch)

	fmt.Println("\n--- Test 3: Worker pool with WaitGroup ---")
	jobs := make(chan int)
	results := make(chan int)
	var wg sync.WaitGroup
	nWorkers := 3
	for w := 1; w <= nWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// produce jobs
	go func() {
		for j := 1; j <= 6; j++ {
			jobs <- j
		}
		close(jobs)
	}()

	// close results when workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		fmt.Println("result:", r)
	}

	fmt.Println("\n--- Test 4: select with timeout ---")
	slow := make(chan string)
	go func() {
		time.Sleep(150 * time.Millisecond)
		slow <- "finished"
	}()

	select {
	case s := <-slow:
		fmt.Println("got:", s)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("timeout")
	}

	fmt.Println("\n--- Test 5: Race condition demo (unsynchronized vs mutex) ---")
	var count int
	var wg2 sync.WaitGroup
	wg2.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg2.Done()
			for j := 0; j < 1000; j++ {
				count++
			}
		}()
	}
	wg2.Wait()
	fmt.Println("unsynchronized count (may be <10000 due to race):", count)

	// synchronized
	count = 0
	var mu sync.Mutex
	wg2.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg2.Done()
			for j := 0; j < 1000; j++ {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}()
	}
	wg2.Wait()
	fmt.Println("synchronized count:", count)

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
