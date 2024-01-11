package scheduled

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	tasks    map[uuid.UUID]*Task
	taskLock sync.Mutex
}

// NewScheduler creates a new instance of the scheduler with an empty initialised slice of tasks.
//
// To use the scheduler, add a task via `scheduler.RegisterTask` or `scheduler.RunOnce`
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make(map[uuid.UUID]*Task),
	}
}

// Register allows for a task to be added within the execution loop of the scheduler
func (s *Scheduler) RegisterTask(t *Task) (uuid.UUID, error) {
	return s.registerTask(t, false)
}

// RunOnce allows a one-time execution of a task directly within the runtime of the scheduler
func (s *Scheduler) RunOnce(t *Task) error {
	_, err := s.registerTask(t, true)
	return err
}

// GetTask returns a task registered under provided uuid.UUID. If the task is not registered, the function returns an error.
func (s *Scheduler) GetTask(idx uuid.UUID) (*Task, error) {
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
func (s *Scheduler) registerTask(t *Task, runOnce bool) (uuid.UUID, error) {
	s.taskLock.Lock()
	defer s.taskLock.Unlock()

	// Go panics with duration <0
	if t.Interval <= time.Duration(0) {
		return uuid.UUID{}, fmt.Errorf("invalid interval=%d", t.Interval)
	}

	// Tasks are identified with Google's UUID, but long-term it would be nice
	// to have internal ID generation to eliminate external dependency
	taskid, _ := uuid.NewUUID()
	if _, ok := s.tasks[taskid]; ok {
		return uuid.UUID{}, fmt.Errorf("taskid=%s is already registered", taskid)
	}

	// Only add the task to the hash-map of tasks if it's not a one-time run.
	if !runOnce {
		s.tasks[taskid] = t
	}

	s.execTask(t, runOnce)
	return taskid, nil
}

// exec is the entrypoint for the execution of the task. It only accepts a task's identifier.
// Tasks are ran in goroutines, which belong to each task respectively.
//
// @TODO: Should exec accept a task pointer to decouple it from the internal array buffer of tasks?
func (s *Scheduler) execTask(task *Task, runOnce bool) {
	go func() {
		time.AfterFunc(time.Until(task.StartTime), func() {
			if err := task.ctx.Err(); err != nil {
				// @TODO: Add IDs to the task itself
				// @TODO: Never print stuff to a console directly
				// A) Return as a slice of errors to the caller
				// B) Expose a logging interface for the caller to have a control over
				// C) Ignore (?)
				fmt.Printf("err: task is cancelled but wanted to be ran\n")
			}

			task.t = time.AfterFunc(task.Interval, func() {
				go task.run()
				defer func() {
					if !runOnce {
						// Reset the internal timer only if the task is supposed to be on a schedule
						task.t.Reset(task.Interval)
					} else {
						// Cancel the context and stop the internal timer for a runOnce task
						task.cancel()
						task.t.Stop()
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
	if s.tasks[taskid].t != nil {
		s.tasks[taskid].t.Stop()
	}
	return nil
}
