/*
Package autodeployment provides a simple, standalone auto-updater for Go applications.

# Overview

This package enables your Go applications to automatically update themselves by
checking a deployment API at regular intervals. It supports SHA256 verification,
time synchronization, and executable permission handling.

# Quick Start

	package main

	import (
		"os"
		"github.com/LucazPlays/AutoDeploymentLib-Go"
	)

	func main() {
		updater := autodeployment.New("your-project-uuid", "your-project-key")
		updater.Start()

		// Your application logic here
		select {}
	}

# Features

  - Automatic update checking at configurable intervals
  - SHA256 hash verification for download integrity
  - Time synchronization with server to prevent unnecessary updates
  - Automatic executable permission handling (0755)
  - Backup creation before update installation
  - Zero external dependencies (Go standard library only)

# Configuration

Set custom API root:

	updater.SetAPIRoot("https://api.insights-api.top/deployment/")

Set custom update interval:

	updater.SetUpdateInterval(60 * time.Second)

# Time Synchronization

The updater synchronizes with the server time to ensure consistent timestamp
comparisons, even if the local system clock is inaccurate.

	info := updater.GetTimeInfo()
	fmt.Printf("Server: %d, Local: %d, Diff: %dms\n",
		info.ServerTime, info.LocalTime, info.TimeDiff)

# API Documentation

For full API documentation, visit: https://pkg.go.dev/github.com/LucazPlays/AutoDeploymentLib-Go
*/
package autodeployment
