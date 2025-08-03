# scheduler Library

The scheduler library provides a simple and flexible job scheduling system for Go applications. It allows you to schedule recurring tasks based on time patterns and manage their execution.

## Installation

```go
import "github.com/getevo/evo/v2/lib/scheduler"
```

## Features

- **Pattern-Based Scheduling**: Schedule jobs using flexible time patterns
- **Concurrent Execution**: Jobs run in separate goroutines
- **Callback Functions**: Handle success, error, and completion events
- **Pause/Resume**: Control job execution with pause functionality
- **Job Management**: Create and manage scheduled jobs

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/scheduler"
    "time"
)

func main() {
    // Register the scheduler (starts the scheduler in a goroutine)
    scheduler.Register()
    
    // Create a job that runs every minute
    job := scheduler.CreateJob(
        "print-time",           // Job ID
        "Mon,*-*-*,*:*:00",     // Run every minute (when seconds are 00)
        func(job *scheduler.Job) error {
            fmt.Printf("Current time: %s\n", time.Now().Format(time.RFC3339))
            return nil
        },
    )
    
    // Start the job
    job.Start()
    
    // Keep the program running
    select {}
}
```

### Using Callbacks

```go
package main

import (
    "errors"
    "fmt"
    "github.com/getevo/evo/v2/lib/scheduler"
    "math/rand"
    "time"
)

func main() {
    // Register the scheduler
    scheduler.Register()
    
    // Create a job that might fail randomly
    job := scheduler.CreateJob(
        "random-task",
        "Mon,*-*-*,*:*:*",  // Run every second
        func(job *scheduler.Job) error {
            // Simulate random failures
            if rand.Intn(3) == 0 {
                return errors.New("random failure")
            }
            fmt.Println("Task completed successfully")
            return nil
        },
    )
    
    // Set up callbacks
    job.OnSuccess = func(job *scheduler.Job) {
        fmt.Println("Success callback: Job completed without errors")
    }
    
    job.OnError = func(job *scheduler.Job, err error) {
        fmt.Printf("Error callback: Job failed with error: %v\n", err)
    }
    
    job.OnFinish = func(job *scheduler.Job) {
        fmt.Println("Finish callback: Job execution finished (regardless of success/failure)")
    }
    
    // Start the job
    job.Start()
    
    // Keep the program running
    select {}
}
```

### Scheduling Patterns

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/scheduler"
)

func main() {
    // Register the scheduler
    scheduler.Register()
    
    // Run every day at midnight
    dailyJob := scheduler.CreateJob(
        "daily-job",
        "Mon,*-*-*,00:00:00",
        func(job *scheduler.Job) error {
            fmt.Println("Daily job running")
            return nil
        },
    )
    
    // Run every Monday at 9:00 AM
    weeklyJob := scheduler.CreateJob(
        "weekly-job",
        "Mon,*-*-*,09:00:00",
        func(job *scheduler.Job) error {
            fmt.Println("Weekly job running")
            return nil
        },
    )
    
    // Run every hour at minute 30
    hourlyJob := scheduler.CreateJob(
        "hourly-job",
        "Mon,*-*-*,*:30:00",
        func(job *scheduler.Job) error {
            fmt.Println("Hourly job running")
            return nil
        },
    )
    
    // Run on the first day of every month at noon
    monthlyJob := scheduler.CreateJob(
        "monthly-job",
        "Mon,*-*-01,12:00:00",
        func(job *scheduler.Job) error {
            fmt.Println("Monthly job running")
            return nil
        },
    )
    
    // Start all jobs
    dailyJob.Start()
    weeklyJob.Start()
    hourlyJob.Start()
    monthlyJob.Start()
    
    // Keep the program running
    select {}
}
```

### Pausing and Resuming Jobs

```go
package main

import (
    "fmt"
    "github.com/getevo/evo/v2/lib/scheduler"
    "time"
)

func main() {
    // Register the scheduler
    scheduler.Register()
    
    // Create a job
    job := scheduler.CreateJob(
        "pausable-job",
        "Mon,*-*-*,*:*:*",  // Run every second
        func(job *scheduler.Job) error {
            fmt.Println("Job running at", time.Now())
            return nil
        },
    )
    
    // Start the job
    job.Start()
    
    // Let it run for a few seconds
    time.Sleep(5 * time.Second)
    
    // Pause the job
    fmt.Println("Pausing job...")
    job.Pause = true
    
    // Wait while paused
    time.Sleep(5 * time.Second)
    
    // Resume the job
    fmt.Println("Resuming job...")
    job.Pause = false
    
    // Keep the program running
    select {}
}
```

## How It Works

The scheduler library uses a pattern-based approach to determine when jobs should run. Each job has an "Every" field that contains a regular expression pattern that is matched against the current time in the format "Mon,2006-01-02,15:04:05".

When you register the scheduler with `scheduler.Register()`, it starts a goroutine that checks every second if any jobs should be executed based on their patterns. If a job's pattern matches the current time and the job is not paused or already running, it executes the job's action function in a separate goroutine.

The `Job` struct contains fields for the job's ID, execution pattern, last run time, action function, and callback functions for handling success, error, and completion events.

The `CreateJob` function creates a new job with the specified ID, pattern, and action function. The pattern uses "*" as a wildcard that matches any alphanumeric character, making it easy to create flexible scheduling patterns.

For more detailed information, please refer to the source code and comments within the library.