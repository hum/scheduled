package scheduled

import (
	"context"
	"fmt"
	"time"

	"github.com/essentialkaos/ek/v12/cron"
)

type TaskOpts struct {
	// Underlying function to be ran within the scheduler.
	Fn TaskFunc

	// In case of an error in Fn, ErrFn will be executed if provided.
	ErrFn TaskErrFunc

	// Offsets the initial startup to a given start time. By default it will start immediately on schedule.
	StartTime time.Time

	// Interval for Fn's execution within the scheduler. This is the function's tick.
	Interval time.Duration

	// Allows the scheduling based on a CRON string. Overrides `Interval`
	Cron string
}

type TaskFunc func() error
type TaskErrFunc func(err error)

type Task struct {
	// Underlying function to be ran within the scheduler
	Fn TaskFunc

	// In case of an error in Fn, ErrFn will be executed if provided
	ErrFn TaskErrFunc

	// Offsets the initial startup to a given start time. By default it will start immediately on schedule.
	StartTime time.Time

	// Interval for Fn's execution within the scheduler. This is the function's tick.
	Interval time.Duration

	// Allows the scheduling based on a CRON string. Overrides `Interval`
	cron *cron.Expr

	ctx    context.Context
	cancel context.CancelFunc

	t *time.Timer
}

func NewTask(opts TaskOpts) *Task {
	var (
		ctx, cancel = context.WithCancel(context.Background())
	)

	// Only set the cron expression if the value is set
	var cronExpr *cron.Expr = nil
	if opts.Cron != "" {
		expr := &opts.Cron

		// Parse the cron expression to validate the task
		c, err := cron.Parse(*expr)
		if err != nil {
			// @TODO: return an error
			panic(err)
		}
		cronExpr = c
	}

	return &Task{
		Fn:        opts.Fn,
		ErrFn:     opts.ErrFn,
		Interval:  opts.Interval,
		StartTime: opts.StartTime,
		cron:      cronExpr,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (t *Task) run() {
	if err := t.Fn(); err != nil {
		if t.ErrFn != nil {
			// While it may not be pretty, it works.
			// Dereferences the pointer into a function type only when ptr != nil
			t.ErrFn(err)
			return
		}
		// @TODO: what is the correct behaviour here?
		fmt.Println("warn: task.Fn returned an error without t.ErrFn set")
		return
	}
}
