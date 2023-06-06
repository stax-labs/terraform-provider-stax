// Package helpers contains helper functions.
package helpers

import (
	"context"
	"fmt"
	"net/http"
)

type HTTPResponse interface {
	StatusCode() int
	Status() string
}

// TaskPoller is a helper to poll an API for the status of an asynchronous task.
// It will call the provided taskFunc to get the latest task status and will continue polling until
// the task completes (succeeds or fails) or an error occurs.
type TaskPoller[T HTTPResponse] struct {
	taskID   string
	taskFunc func() (T, error)
	lastResp T
	err      error // Sticky error.
}

// NewTaskPoller creates a new TaskPoller.
// taskID is the ID of the asynchronous task to poll.
// taskFunc is a function that will be called to get the latest status of the task. It should return
// a HTTPResponse and an error.
func NewTaskPoller[T HTTPResponse](taskID string, taskFunc func() (T, error)) *TaskPoller[T] {
	return &TaskPoller[T]{taskID: taskID, taskFunc: taskFunc}
}

// Err returns the first error encountered while polling.
func (s *TaskPoller[T]) Err() error {
	return s.err
}

// Resp returns the latest response from calling taskFunc.
func (s *TaskPoller[T]) Resp() T {
	return s.lastResp
}

// Poll calls taskFunc to get the latest task status. It will continue polling until the task completes
// (succeeds or fails) or an error occurs.
// Returns true if polling should continue, false otherwise.
func (s *TaskPoller[T]) Poll(ctx context.Context) bool {

	s.lastResp, s.err = s.taskFunc()
	if s.err != nil {
		return false
	}
	// TODO: Implement some basic error count back off
	if s.lastResp.StatusCode() != http.StatusOK {

		s.err = fmt.Errorf("request failed, returned non 200 status: %s", s.lastResp.Status())
		return false
	}

	return true
}
