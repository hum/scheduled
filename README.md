# scheduled
**scheduled** implements an in-memory scheduler to run functions in a time loop. It also supports running functions with a CRON syntax scheduling.

---

## Usage
### Get
```bash
> go get github.com/hum/scheduled
```
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
You can also bound the `Interval` with `MaxIter` to only allow N intervals to pass.
```go
task := scheduled.NewTask(scheduled.TaskOpts{
  Fn: func() error {
    fmt.Println("Hello from task!")
    return nil
  },
  Interval: 10*time.Second,
  // Stops executing after 10 runs
  MaxIter: 10,
})
```

### With start time and end time
You can also specify `StartTime` and/or `EndTime` to time bound the task's execution. `StartTime` allows the scheduler to wait until the specified time to execute the function, while `EndTime` is used with interval or CRON to end the execution of the task after a certain period of time passes.
```go
task := scheduled.NewTask(scheduled.TaskOpts{
  Fn: func() error {
    fmt.Println("Hello from task!")
    return nil
  },
  Interval: 10*time.Second,
  // Starts the function in 5 hours, then runs it every 10 seconds
  StartTime: time.Now().Add(5*time.Hour),
  // Stops executing after 24 hours
  EndTime: time.Now().Add(24*time.Hour),
})
```

### With CRON
A task supports a CRON syntax to act as the interval function. Use `Cron` to pass in a CRON string.


NOTE: `Cron` has a priority over `Interval`. If you pass both, `Cron` will be used.
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
