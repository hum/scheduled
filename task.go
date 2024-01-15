package scheduled

import (
	"context"
	"fmt"
	"time"

	"github.com/essentialkaos/ek/v12/cron"
	"github.com/google/uuid"
)

type TaskOpts struct {
	// Underlying function to be ran within the scheduler.
	Fn TaskFunc

	// In case of an error in Fn, ErrFn will be executed if provided.
	ErrFn TaskErrFunc

	// Offsets the initial startup to a given start time. By default it will start immediately on schedule.
	StartTime time.Time

	// Restricts the time boundary of the runtime to an end date. By default it is unbound.
	EndTime time.Time

	// Interval for Fn's execution within the scheduler. This is the function's tick.
	Interval time.Duration

	// Allows the scheduling based on a CRON string. Overrides `Interval`
	Cron string
}

type TaskFunc func() error
type TaskErrFunc func(err error)

type task struct {
	// Task identifier is used to be able to get and stop tasks already registered in the scheduler
	ID uuid.UUID

	// Underlying function to be ran within the scheduler
	Fn TaskFunc

	// In case of an error in Fn, ErrFn will be executed if provided
	ErrFn TaskErrFunc

	// Offsets the initial startup to a given start time. By default it will start immediately on schedule.
	StartTime time.Time

	// Restricts the time boundary of the runtime to an end date. By default it is unbound.
	EndTime *time.Time

	// Interval for Fn's execution within the scheduler. This is the function's tick.
	Interval time.Duration

	// Allows the scheduling based on a CRON string. Overrides `Interval`
	cron *cron.Expr

	ctx    context.Context
	cancel context.CancelFunc

	timer *time.Timer
}

func NewTask(opts TaskOpts) *task {
	// If the caller has picked neither CRON nor interval, then we have no idea how to schedule the task
	if opts.Cron == "" && opts.Interval <= 0 {
		panic("neither cron nor interval is set")
	}

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

	var (
		// uuid.NewUUID() only returns an (uuid, error) for historical reasons.
		// It never returns an error anymore, so can be safely ignored.
		// [Issue](https://github.com/google/uuid/issues/63)
		taskid, _ = uuid.NewUUID()

		// Internal context for the task. Used for cancellation.
		ctx, cancel = context.WithCancel(context.Background())

		endTime *time.Time
	)

	// If EndTime is not set, it defaults to 0001-01-01 00:00:00 +0000 UTC
	if !opts.EndTime.IsZero() {
		// Only allow end time if it is in the future
		if opts.EndTime.Before(time.Now()) {
			panic(fmt.Sprintf("task was scheduled to end in the past: %s", opts.EndTime.String()))
		}
		endTime = &opts.EndTime
	}

	return &task{
		ID:        taskid,
		Fn:        opts.Fn,
		ErrFn:     opts.ErrFn,
		Interval:  opts.Interval,
		StartTime: opts.StartTime,
		EndTime:   endTime,
		cron:      cronExpr,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (t *task) run() {
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

// resetTimer sets the next tick for the task to run with its internal timer.
func (t *task) resetTimer() {
	var next = t.Interval
	if t.cron != nil {
		// CRON has a precedence over interval
		next = time.Until(t.cron.Next())
	}
	t.timer.Reset(next)
}
