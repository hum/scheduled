package scheduled_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hum/scheduled"
	"github.com/stretchr/testify/require"
)

func TestSchedulerAddsTask(t *testing.T) {
	task := scheduled.NewTask(scheduled.TaskOpts{
		Fn: func() error {
			fmt.Println("hello, world!")
			return nil
		},
		// Large enough interval for the whole test to finish so that we do not print anything to the console
		Interval: 15 * time.Second,
	})

	scheduler := scheduled.NewScheduler()
	taskid, err := scheduler.RegisterTask(task)
	require.NoError(t, err)

	registeredTask, err := scheduler.GetTask(taskid)
	require.NoError(t, err)
	require.Equal(t, task, registeredTask)

	err = scheduler.RemoveTask(taskid)
	require.NoError(t, err)
}

func TestSchedulerRemovesTask(t *testing.T) {
	task := scheduled.NewTask(scheduled.TaskOpts{
		Fn: func() error {
			fmt.Println("hello, world!")
			return nil
		},
		Interval: time.Second,
	})

	scheduler := scheduled.NewScheduler()
	taskid, err := scheduler.RegisterTask(task)
	require.NoError(t, err)

	err = scheduler.RemoveTask(taskid)
	require.NoError(t, err)

	_, err = scheduler.GetTask(taskid)
	require.Error(t, err)
}

func TestSchedulerExecutesTaskAtLeastOnce(t *testing.T) {
	var finishedChan = make(chan bool, 1)

	task := scheduled.NewTask(scheduled.TaskOpts{
		Fn: func() error {
			finishedChan <- true
			return nil
		},
		Interval: time.Second,
	})

	scheduler := scheduled.NewScheduler()
	_, err := scheduler.RegisterTask(task)
	require.NoError(t, err)

test_loop:
	for timeout := time.After(5 * time.Second); ; {
		select {
		case <-timeout:
			t.Fatalf("test timed out before the task ran")
		case <-finishedChan:
			break test_loop
		}
	}
}

func TestSchedulerExecutesTaskOneTime(t *testing.T) {
	var finishedChan = make(chan bool, 10)

	task := scheduled.NewTask(scheduled.TaskOpts{
		Fn: func() error {
			finishedChan <- true
			return nil
		},
		Interval: time.Second,
	})

	scheduler := scheduled.NewScheduler()
	err := scheduler.RunOnce(task)
	require.NoError(t, err)

	var count = 0

test_loop:
	for timeout := time.After(5 * time.Second); ; {
		select {
		case <-timeout:
			if count == 0 || count > 1 {
				t.Fatalf("task did not execute only once, count=%d", count)
			}
			break test_loop
		case <-finishedChan:
			count++
		}
	}
}

func TestSchedulerExecutesTaskMultipleTimes(t *testing.T) {
	var finishedChan = make(chan bool, 10)

	task := scheduled.NewTask(scheduled.TaskOpts{
		Fn: func() error {
			finishedChan <- true
			return nil
		},
		Interval: time.Second,
	})

	scheduler := scheduled.NewScheduler()
	idx, err := scheduler.RegisterTask(task)
	require.NoError(t, err)

	var count = 0

test_loop:
	for timeout := time.After(5 * time.Second); ; {
		select {
		case <-timeout:
			if count <= 1 {
				t.Fatalf("task did not execute multiple times, count=%d", count)
			}
			break test_loop
		case <-finishedChan:
			count++
		}
	}

	err = scheduler.RemoveTask(idx)
	require.NoError(t, err)
}
