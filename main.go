package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"os"
"github.com/fatih/color"

)

// PortScanResult stores information about an open port
// It includes the target IP/hostname, port number, status, and optional banner response

type PortScanResult struct {
	Target string `json:"target"`
	Port   int    `json:"port"`
	Status string `json:"status"`
	Banner string `json:"banner"`
}

func (p PortScanResult) MarshalJSON() ([]byte, error) {
	if p.Banner == "" {
		p.Banner = "No banner"
	}
	type Alias PortScanResult
	return json.Marshal((Alias)(p))
}

// worker handles scanning tasks concurrently
// It attempts to connect to a given address and optionally grabs the service banner
func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, openPorts *[]PortScanResult, mu *sync.Mutex, totalPorts, scanned *int) {
	defer wg.Done()
	maxRetries := 1

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	for addr := range tasks {
		var success bool
		var banner string
		parts := strings.Split(addr, ":")
		port, _ := strconv.Atoi(parts[1])
		target := parts[0]

		for i := 0; i < maxRetries; i++ {
			conn, err := dialer.Dial("tcp", addr)
			if err == nil {
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				if err == nil && n > 0 {
					banner = strings.TrimSpace(string(buffer[:n]))
				}
				conn.Close()

				mu.Lock()
				*openPorts = append(*openPorts, PortScanResult{Target: target, Port: port, Status: "open", Banner: banner})
				mu.Unlock()

				fmt.Printf("\r%s %s:%d %s\n", green("[OPEN]"), target, port, banner)
				success = true
				break
			}
			time.Sleep(time.Duration(1<<i) * time.Second)
		}

		if !success {
			fmt.Printf("\r%s %s:%d\n", red("[CLOSED]"), target, port)
		}

		// Update scanning progress
		mu.Lock()
*scanned++
fmt.Fprintf(os.Stdout, "\r%s Scanning port %d/%d...", cyan("[*]"), *scanned, *totalPorts)
mu.Unlock()
	}
}


func main() {
	// Command-line flags for user input
	target := flag.String("target", "", "Specify a single target IP or hostname")
	targets := flag.String("targets", "", "Comma-separated list of target IPs or hostnames")
	startPort := flag.Int("start-port", 1, "Starting port number")
	endPort := flag.Int("end-port", 22, "Ending port number")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	timeout := flag.Int("timeout", 5, "Connection timeout in seconds")
	jsonOutput := flag.Bool("json", false, "Output results in JSON format")
	portsList := flag.String("ports", "", "Comma-separated list of specific ports to scan (e.g., 22,80,443)")

	flag.Parse()

	// Determine target list based on user input
	// Determine target list based on user input
var targetList []string

if *target != "" {
    targetList = append(targetList, *target)
}

if *targets != "" {
    additionalTargets := strings.Split(*targets, ",")
    for _, t := range additionalTargets {
        trimmed := strings.TrimSpace(t)
        if trimmed != "" {
            targetList = append(targetList, trimmed)
        }
    }
}

if len(targetList) == 0 {
    fmt.Println("Error: No targets specified. Use -target or -targets flag.")
    return
}


	// Parse specific ports (if provided)
	portSet := make(map[int]bool)
	for p := *startPort; p <= *endPort; p++ {
		portSet[p] = true
	}
	if *portsList != "" {
		for _, p := range strings.Split(*portsList, ",") {
			port, err := strconv.Atoi(strings.TrimSpace(p))
			if err == nil {
				portSet[port] = true
			}
		}
	}
	var portRange []int
	for port := range portSet {
		portRange = append(portRange, port)
	}

	// Initialize scanning process
	startTime := time.Now()
	var wg sync.WaitGroup
	tasks := make(chan string, len(targetList)*len(portRange))
	var openPorts []PortScanResult
	var mu sync.Mutex
	dialer := net.Dialer{Timeout: time.Duration(*timeout) * time.Second}

	totalPorts := len(targetList) * len(portRange)
	scanned := 0

	// Launch worker goroutines for concurrent scanning
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer, &openPorts, &mu, &totalPorts, &scanned)
	}

	// Queue scan tasks for each target and port combination
	for _, target := range targetList {
		for _, port := range portRange {
			tasks <- net.JoinHostPort(target, strconv.Itoa(port))
		}
	}

	close(tasks) // Close task channel to signal workers that no more tasks are coming
	wg.Wait()    // Wait for all workers to finish

	// Compute scan duration
	duration := time.Since(startTime)
	fmt.Println("\n\n=== Scan Summary ===")
fmt.Printf("Targets Scanned: %d\n", len(targetList))
fmt.Printf("Ports Scanned: %d\n", totalPorts)
fmt.Printf("Open Ports Found: %d\n", len(openPorts))
fmt.Printf("Total Duration: %v\n", duration)


	// Output results in JSON format if requested
	if *jsonOutput {
		jsonData, err := json.MarshalIndent(openPorts, "", "  ")
		if err != nil {
			fmt.Println("Error generating JSON output")
		} else {
			if len(openPorts) == 0 {
				fmt.Println("[]") // Print empty JSON array if no open ports found
			} else {
				fmt.Println(string(jsonData))
			}
		}
	}
}
