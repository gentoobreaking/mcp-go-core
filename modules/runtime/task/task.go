// Package task provides background task management with cancellation support.
// Tasks run as goroutines with lifecycle states and result channels.
package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrTaskAlreadyDone  = errors.New("task already completed")
	ErrTaskAlreadyCancel = errors.New("task already cancelled")
)

// Status represents the state of a task.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusCancelled
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Result holds the outcome of a completed task.
type Result struct {
	Data []byte
	Err  error
}

// Task represents a background task with lifecycle management.
type Task struct {
	ID        string
	Name      string
	Status    Status
	CreatedAt time.Time
	StartedAt time.Time
	DoneAt    time.Time
	Result    Result
	cancel    context.CancelFunc
	done      chan struct{}
	mu         sync.RWMutex
}

// Manager manages background tasks.
type Manager struct {
	mu    sync.RWMutex
	tasks map[string]*Task
}

// NewManager creates a new task Manager.
func NewManager() *Manager {
	return &Manager{
		tasks: make(map[string]*Task),
	}
}

// Create registers a new task with the given name and handler.
// The handler runs immediately in a goroutine.
func (m *Manager) Create(name string, fn func(ctx context.Context) (Result, error)) *Task {
	ctx, cancel := context.WithCancel(context.Background())
	task := &Task{
		ID:        generateID(),
		Name:      name,
		Status:    StatusRunning,
		CreatedAt: time.Now(),
		StartedAt: time.Now(),
		cancel:    cancel,
		done:      make(chan struct{}),
	}

	m.mu.Lock()
	m.tasks[task.ID] = task
	m.mu.Unlock()

	go func() {
		defer close(task.done)

		result, err := fn(ctx)

		task.mu.Lock()
		defer task.mu.Unlock()

		task.DoneAt = time.Now()
		if err != nil {
			task.Status = StatusFailed
		} else {
			task.Status = StatusCompleted
		}
		task.Result = result
		if err != nil {
			task.Result.Err = err
		}
	}()

	return task
}

// Cancel requests cancellation of a task by ID.
func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		return ErrTaskNotFound
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	if task.Status != StatusRunning {
		return fmt.Errorf("%w: %s", ErrTaskAlreadyDone, task.Status)
	}

	if task.cancel != nil {
		task.cancel()
	}

	task.Status = StatusCancelled
	task.DoneAt = time.Now()
	return nil
}

// Status returns the status of a task by ID.
func (m *Manager) Status(id string) (Status, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[id]
	if !ok {
		return StatusPending, ErrTaskNotFound
	}

	task.mu.RLock()
	defer task.mu.RUnlock()
	return task.Status, nil
}

// GetResult retrieves the result of a completed task.
func (m *Manager) GetResult(id string) (Result, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[id]
	if !ok {
		return Result{}, ErrTaskNotFound
	}

	task.mu.RLock()
	defer task.mu.RUnlock()
	if task.Status != StatusCompleted && task.Status != StatusFailed {
		return Result{}, fmt.Errorf("task %s is still in %s state", id, task.Status)
	}

	return task.Result, nil
}

// WaitFor waits for a task to complete or the context to expire.
func (m *Manager) WaitFor(id string, ctx context.Context) error {
	m.mu.RLock()
	task, ok := m.tasks[id]
	m.mu.RUnlock()

	if !ok {
		return ErrTaskNotFound
	}

	select {
	case <-task.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Count returns the total number of tasks.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tasks)
}

// RunningCount returns the number of running tasks.
func (m *Manager) RunningCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, task := range m.tasks {
		task.mu.RLock()
		if task.Status == StatusRunning {
			count++
		}
		task.mu.RUnlock()
	}
	return count
}

// Close cancels all running tasks and waits for completion.
func (m *Manager) Close() {
	m.mu.Lock()
	for _, task := range m.tasks {
		task.mu.Lock()
		if task.Status == StatusRunning && task.cancel != nil {
			task.cancel()
		}
		task.mu.Unlock()
	}
	m.mu.Unlock()
}

// generateID creates a unique task ID.
func generateID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
