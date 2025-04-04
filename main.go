package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

var openPorts int
var mu sync.Mutex

func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer) {
	defer wg.Done()
	maxRetries := 3
	for addr := range tasks {
		var success bool
		for i := range maxRetries {
			conn, err := dialer.Dial("tcp", addr)
			if err == nil {
				conn.Close()
				fmt.Printf("Connection to %s was successful\n", addr)
				mu.Lock()
				openPorts++
				mu.Unlock()
				success = true
				break
			}
			backoff := time.Duration(1<<i) * time.Second
			fmt.Printf("Attempt %d to %s failed. Waiting %v...\n", i+1, addr, backoff)
			time.Sleep(backoff)
		}
		if !success {
			fmt.Printf("Failed to connect to %s after %d attempts\n", addr, maxRetries)
		}
	}
}

func main() {
	target := flag.String("target", "scanme.nmap.org", "Target IP address or hostname")
	startPort := flag.Int("start-port", 1, "Starting port range")
	endPort := flag.Int("end-port", 1024, "Ending port range")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	flag.Parse()

	startTime := time.Now()
	var wg sync.WaitGroup
	tasks := make(chan string, 100)
	dialer := net.Dialer{
		Timeout: 5 * time.Second,
	}

	for i := 1; i <= *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer)
	}

	totalPorts := 0
	for p := *startPort; p <= *endPort; p++ {
		port := strconv.Itoa(p)
		address := net.JoinHostPort(*target, port)
		tasks <- address
		totalPorts++
	}
	close(tasks)
	wg.Wait()

	duration := time.Since(startTime)
	fmt.Println("\n--- Scan Summary ---")
	fmt.Printf("Total ports scanned: %d\n", totalPorts)
	fmt.Printf("Number of open ports: %d\n", openPorts)
	fmt.Printf("Time taken: %v\n", duration)
}
