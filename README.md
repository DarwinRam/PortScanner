# PortScanner

##  Description
This is a high-performance, concurrent TCP port scanner written in Go. It allows scanning a single target or multiple hosts, specifying custom port ranges or lists, and fetching service banners. The tool is designed for efficiency, leveraging Goroutines for parallel scanning.

## 🚀 How to Build and Run

### ** Prerequisites**
- Install [Go](https://go.dev/doc/install) 

### ** Clone the Repository**

### packages
```For color package from github use in terminal:
go get github.com/fatih/color
```


### **  Run the Scanner**
#### Scan default ports (1–1024) on localhost
```bash
 go run main.go  -target 127.0.0.1
```

#### Scan specific ports on a target
```bash
 go run main.go  -target scanme.nmap.org -ports 22,80,443
```

#### Scan a range of ports on multiple targets
```bash
 go run main.go  -targets 192.168.1.1,192.168.1.2 -start-port 20 -end-port 100
```

#### Run with JSON output
```bash
 go run main.go  -target 127.0.0.1 -ports 22,80 -json
```

### **Text Output**
```plaintext
Scanning port 2/100...
[OPEN] 127.0.0.1:22 SSH-2.0-OpenSSH_8.6
[OPEN] 127.0.0.1:80 HTTP/1.1 200 OK
[CLOSED] 127.0.0.1:443
=== Scan Summary ===
Targets Scanned: 1
Ports Scanned: 3
Open Ports: 2
Scan Duration: 1.2s
```

### **JSON Output**
```json
[
  {
    "target": "127.0.0.1",
    "port": 22,
    "status": "open",
    "banner": "SSH-2.0-OpenSSH_8.6"
  },
  {
    "target": "127.0.0.1",
    "port": 80,
    "status": "open",
    "banner": "HTTP/1.1 200 OK"
  }
]
```