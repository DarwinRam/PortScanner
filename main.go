// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go with -target, -targets, -start-port, -end-port, -workers, and -timeout flags. It also performs banner grabbing and shows progress.

package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var openPorts int
var mu sync.Mutex
var progressCounter int

func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, totalPorts int) {
	defer wg.Done()
	maxRetries := 3
	for addr := range tasks {
		mu.Lock()
		progressCounter++
		fmt.Printf("Scanning port %d/%d: %s\n", progressCounter, totalPorts, addr)
		mu.Unlock()

		var success bool
		for i := range maxRetries {
			conn, err := dialer.Dial("tcp", addr)
			if err == nil {
				fmt.Printf("Connection to %s was successful\n", addr)
				mu.Lock()
				openPorts++
				mu.Unlock()
				success = true

				// Attempt to grab banner
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err == nil && n > 0 {
					fmt.Printf("Banner from %s: %s\n", addr, string(buffer[:n]))
				} else {
					fmt.Printf("No banner received from %s\n", addr)
				}
				conn.Close()

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
	target := flag.String("target", "", "Single target IP address or hostname")
	targets := flag.String("targets", "scanme.nmap.org", "Comma-separated list of target IP addresses or hostnames")
	startPort := flag.Int("start-port", 1, "Starting port range")
	endPort := flag.Int("end-port", 22, "Ending port range")
	workers := flag.Int("workers", 200, "Number of concurrent workers")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	flag.Parse()

	// Use the single target if provided, otherwise fall back to the targets list
	var targetList []string
	if *target != "" {
		targetList = append(targetList, *target)
	} else {
		targetList = strings.Split(*targets, ",")
	}

	if len(targetList) == 0 {
		fmt.Println("Error: No target specified. Use -target or -targets.")
		return
	}

	totalPorts := (*endPort - *startPort + 1) * len(targetList)
	startTime := time.Now()

	var wg sync.WaitGroup
	tasks := make(chan string, 100)
	dialer := net.Dialer{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	for i := 1; i <= *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer, totalPorts)
	}

	for _, target := range targetList {
		target = strings.TrimSpace(target)
		for p := *startPort; p <= *endPort; p++ {
			port := strconv.Itoa(p)
			address := net.JoinHostPort(target, port)
			tasks <- address
		}
	}
	close(tasks)
	wg.Wait()

	duration := time.Since(startTime)
	fmt.Println("\n--- Scan Summary ---")
	fmt.Printf("Total ports scanned: %d\n", totalPorts)
	fmt.Printf("Number of open ports: %d\n", openPorts)
	fmt.Printf("Time taken: %v\n", duration)
}
