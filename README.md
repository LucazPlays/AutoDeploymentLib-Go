# AutoDeploymentLib-Go

A simple, standalone auto-updater for Go applications.

## Installation

```bash
go get github.com/insights-autodeployment/autoupdater
```

## Usage

```go
package main

import (
	"os"
	"time"
	"github.com/insights-autodeployment/autoupdater"
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
	
	// Your application code here
	// ...
	
	select {}
}
```

## API

```go
updater := autodeployment.New(uuid, key)
updater.SetAPIRoot("https://api.insights-api.top/deployment/")
updater.SetUpdateInterval(30 * time.Second)
updater.Start()
updater.Stop()
```

## License

MIT
