package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

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
	flag.Parse()

	var wg sync.WaitGroup
	tasks := make(chan string, 100)
	dialer := net.Dialer{
		Timeout: 5 * time.Second,
	}

	workers := 100
	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer)
	}

	for p := *startPort; p <= *endPort; p++ {
		port := strconv.Itoa(p)
		address := net.JoinHostPort(*target, port)
		tasks <- address
	}
	close(tasks)
	wg.Wait()
}