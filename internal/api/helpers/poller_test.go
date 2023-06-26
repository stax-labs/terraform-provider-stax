package helpers

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskPoller(t *testing.T) {
	t.Run("polls successfully", func(t *testing.T) {
		taskFunc := func() (HTTPResponse, error) {
			return &mockHTTPResponse{statusCode: http.StatusOK}, nil
		}
		tp := NewTaskPoller(taskFunc)

		shouldContinue := tp.Poll(context.Background())
		assert.True(t, shouldContinue)
		assert.Nil(t, tp.Err())
		assert.NotNil(t, tp.Resp())
	})

	t.Run("stops polling on error", func(t *testing.T) {
		taskFunc := func() (HTTPResponse, error) {
			return nil, fmt.Errorf("error")
		}
		tp := NewTaskPoller(taskFunc)

		shouldContinue := tp.Poll(context.Background())
		assert.False(t, shouldContinue)
		assert.NotNil(t, tp.Err())
		assert.Nil(t, tp.Resp())
	})

	t.Run("stops polling on non-200 response", func(t *testing.T) {
		taskFunc := func() (HTTPResponse, error) {
			return &mockHTTPResponse{statusCode: http.StatusBadRequest}, nil
		}
		tp := NewTaskPoller(taskFunc)

		shouldContinue := tp.Poll(context.Background())
		assert.False(t, shouldContinue)
		assert.NotNil(t, tp.Err())
		assert.Equal(t, http.StatusBadRequest, tp.Resp().StatusCode())
	})
}

type mockHTTPResponse struct {
	statusCode int
}

func (m *mockHTTPResponse) StatusCode() int {
	return m.statusCode
}

func (m *mockHTTPResponse) Status() string {
	return http.StatusText(m.statusCode)
}
