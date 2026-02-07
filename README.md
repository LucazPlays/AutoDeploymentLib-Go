# AutoDeploymentLib-Go

[![Go Reference](https://pkg.go.dev/badge/github.com/LucazPlays/AutoDeploymentLib-Go.svg)](https://pkg.go.dev/github.com/LucazPlays/AutoDeploymentLib-Go)
[![Go Version](https://img.shields.io/badge/Go-1.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

A simple, standalone auto-updater for Go applications with SHA256 verification and time synchronization.

## Installation

```bash
go get github.com/LucazPlays/AutoDeploymentLib-Go
```

## Quick Start

```go
package main

import (
	"os"
	"time"
	"github.com/LucazPlays/AutoDeploymentLib-Go"
)

func main() {
	updater := autodeployment.New("your-project-uuid", "your-project-key")
	
	// Optional configuration
	updater.SetAPIRoot("https://api.insights-api.top/deployment/")
	updater.SetUpdateInterval(30 * time.Second)
	
	// Start the updater
	if err := updater.Start(); err != nil {
		panic(err)
	}
	
	// Your application logic here
	// ...
	
	select {}
}
```

## Features

- **Automatic Updates** - Periodically checks for new releases from the deployment API
- **SHA256 Verification** - Ensures download integrity with cryptographic hash verification
- **Time Synchronization** - Syncs with server time to prevent unnecessary updates
- **Executable Permissions** - Automatically sets 0755 permissions after update
- **Backup Creation** - Creates .bak backup before installing updates
- **Zero Dependencies** - Uses only the Go standard library

## API Documentation

### Updater

```go
// Create a new updater
updater := autodeployment.New(uuid, key)

// Configure API root (default: https://api.insights-api.top/deployment/)
updater.SetAPIRoot("https://api.insights-api.top/deployment/")

// Configure update interval (default: 30 seconds)
updater.SetUpdateInterval(30 * time.Second)

// Start the updater
updater.Start()

// Stop the updater
updater.Stop()
```

### Time Synchronization

```go
// Sync time with server (called automatically by Start)
updater.SyncTime()

// Get all timing information for debugging
info := updater.GetTimeInfo()
// info.ServerTime        - Server time (Unix ms)
// info.LocalTime         - Local time (Unix ms)
// info.TimeDiff          - Server - Local (ms)

// Get individual values
serverTime := updater.GetServerTime()
localTime := updater.GetLocalTime()
timeDiff := updater.GetTimeDiff()
```

## License

MIT License - see [LICENSE](LICENSE) for details.
