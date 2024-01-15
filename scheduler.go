package scheduled

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	tasks    map[uuid.UUID]*task
	taskLock sync.Mutex
}

// NewScheduler creates a new instance of the scheduler with an empty initialised slice of tasks.
//
// To use the scheduler, add a task via `scheduler.RegisterTask` or `scheduler.RunOnce`
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make(map[uuid.UUID]*task),
	}
}

// RegisterTask allows for a task to be added within the execution loop of the scheduler
func (s *Scheduler) RegisterTask(t *task) error {
	return s.registerTask(t, false)
}

// RunOnce allows a one-time execution of a task directly within the runtime of the scheduler
func (s *Scheduler) RunOnce(t *task) error {
	return s.registerTask(t, true)
}

// GetTask returns a task registered under provided uuid.UUID. If the task is not registered, the function returns an error.
func (s *Scheduler) GetTask(idx uuid.UUID) (*task, error) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	task, ok := s.tasks[idx]
	if !ok {
		return nil, fmt.Errorf("could not find a task with id=%s", idx)
	}
	return task, nil
}

// RemoveTask removes a given task from the scheduler while also stopping its execution if it were already scheduled
func (s *Scheduler) RemoveTask(idx uuid.UUID) error {
	err := s.stopTask(idx)
	if err != nil {
		return err
	}

	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	delete(s.tasks, idx)
	return nil
}

// Stop removes all scheduled tasks from the scheduler
func (s *Scheduler) Stop() {
	for taskid := range s.tasks {
		err := s.RemoveTask(taskid)
		if err != nil {
			// @TODO: Never print stuff to a console directly
			// A) Return as a slice of errors to the caller
			// B) Expose a logging interface for the caller to have a control over
			// C) Ignore (?)
			fmt.Printf("Err: Stop() could not remove task from the scheduler, got err=%s", err)
			continue
		}
	}
}

// registerTask validates task's correctness before adding it to the group of tasks.
// Tasks can run either on a schedule, or be executed only once within the context of the scheduler via the parameter `runOnce`.
func (s *Scheduler) registerTask(t *task, runOnce bool) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	if _, ok := s.tasks[t.ID]; ok {
		return fmt.Errorf("taskid=%s is already registered", t.ID)
	}

	s.tasks[t.ID] = t
	s.execTask(t, runOnce)
	return nil
}

// execTask is the entrypoint for the execution of the task. It only accepts a task's identifier.
// Tasks are ran in goroutines, which belong to each task respectively.
//
// @TODO: Should exec accept a task pointer to decouple it from the internal array buffer of tasks?
func (s *Scheduler) execTask(task *task, runOnce bool) {
	go func() {
		time.AfterFunc(time.Until(task.StartTime), func() {
			if err := task.ctx.Err(); err != nil {
				// @TODO: Never print stuff to a console directly
				// A) Return as a slice of errors to the caller
				// B) Expose a logging interface for the caller to have a control over
				// C) Ignore (?)
				fmt.Printf("err: task=%s is cancelled but wanted to be ran\n", task.ID)

				// Make sure to also stop the tick timer
				if task.timer != nil {
					task.timer.Stop()
				}
				return
			}

			// Default tick is the task's interval
			var tick time.Duration = task.Interval

			// If the task type is CRON, use the next CRON time as the tick
			if task.cron != nil {
				tick = time.Until(task.cron.Next())
			}

			task.timer = time.AfterFunc(tick, func() {
				// Make sure to check if the end time has not been exceeded
				if task.EndTime != nil && time.Now().After(*task.EndTime) {
					s.RemoveTask(task.ID)
					return
				}

				go task.run()
				defer func() {
					if !runOnce {
						task.resetTimer()
					} else {
						s.RemoveTask(task.ID)
					}
				}()
			})
		})
	}()
}

func (s *Scheduler) stopTask(taskid uuid.UUID) error {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	// @TODO: Does it matter if we try to delete a task which does not exist?
	// Should this just be a no-op for the caller?
	if _, ok := s.tasks[taskid]; !ok {
		return fmt.Errorf("task=%s is not registered", taskid)
	}

	// Cancels the context of the task
	// @TODO: maybe have a function on the task itself to cancel gracefully?
	s.tasks[taskid].cancel()

	// Stops the internal timer of the task
	if s.tasks[taskid].timer != nil {
		s.tasks[taskid].timer.Stop()
	}
	return nil
}
