package main

import (
	"fmt"
	"time"
)

func generator(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= n; i++ {
			out <- i
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

func main() {
	fmt.Println("=== MANUAL TEST MATERI 17: CHANNEL ===\n")

	fmt.Println("--- Test 1: Simple send/receive ---")
	ch := make(chan int)
	go func() { ch <- 1 }()
	fmt.Println("received:", <-ch)

	fmt.Println("\n--- Test 2: Close and range ---")
	ch2 := make(chan int)
	go func() {
		ch2 <- 1
		ch2 <- 2
		ch2 <- 3
		close(ch2)
	}()
	for x := range ch2 {
		fmt.Println("range:", x)
	}

	fmt.Println("\n--- Test 3: Buffered channel ---")
	buf := make(chan int, 2)
	buf <- 10
	buf <- 20
	fmt.Println("buf recv:", <-buf, <-buf)

	fmt.Println("\n--- Test 4: Select fan-in ---")
	a := make(chan string)
	b := make(chan string)
	go func() { time.Sleep(20 * time.Millisecond); a <- "from a" }()
	go func() { time.Sleep(10 * time.Millisecond); b <- "from b" }()
	for i := 0; i < 2; i++ {
		select {
		case s := <-a:
			fmt.Println("got:", s)
		case s := <-b:
			fmt.Println("got:", s)
		case <-time.After(30 * time.Millisecond):
			fmt.Println("timeout")
		}
	}

	fmt.Println("\n--- Test 5: Pipeline (generator->sq) ---")
	for n := range sq(generator(5)) {
		fmt.Println("sq:", n)
	}

	fmt.Println("\n--- Test 6: Nil channel blocks select ---")
	var nilCh chan int
	select {
	case <-nilCh:
		fmt.Println("unexpected")
	case <-time.After(20 * time.Millisecond):
		fmt.Println("nil channel selective timeout")
	}

	fmt.Println("\n=== SEMUA TEST SELESAI ===")
}
