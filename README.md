# scheduled
**scheduled** implements an in-memory scheduler to run functions in a time loop. It also supports running functions with a CRON syntax scheduling.

---

## Usage
### Basic
Allows for a creation of simple functions, which get executed every `Interval` tick.
```go
import "github.com/hum/scheduled"

// Create a new task via scheduled.NewTask
task := scheduled.NewTask(scheduled.TaskOpts{
  // The actual function to be ran
  Fn: func() error {
    fmt.Println("Hello from task!")
    return nil
  },
  // How often should the function be ran
  Interval: 10*time.Second,
})

// Initialise a new scheduler
scheduler := scheduled.NewScheduler()
defer scheduler.Stop()

// Add the task to the scheduler
scheduler.RegisterTask(task)
```

### With start time
You can also specify `Task.StartTime` to be in the future, therefore delaying the execution of the function until the specified time.
```go
task := scheduled.NewTask(scheduled.TaskOpts{
  Fn: func() error {
    fmt.Println("Hello from task!")
    return nil
  },
  Interval: 10*time.Second,
  // Starts the function in 5 hours, then runs it every 10 seconds
  StartTime: time.Now().Add(5*time.Hour),
})
```

### With CRON
A task supports a CRON syntax to act as the interval function. Use `Task.Cron` to pass in a CRON string.
```go
task := scheduled.NewTask(scheduled.TaskOpts{
  Fn: func() error {
    fmt.Println("Hello from task!")
    return nil
  },
  // Run every minute
  Cron: "* * * * *",
})
```
