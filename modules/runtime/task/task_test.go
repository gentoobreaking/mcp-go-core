package task

import (
	"context"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
	if m.Count() != 0 {
		t.Fatalf("expected 0 tasks, got %d", m.Count())
	}
}

func TestCreateAndStatus(t *testing.T) {
	m := NewManager()
	defer m.Close()

	task := m.Create("test-task", func(ctx context.Context) (Result, error) {
		return Result{Data: []byte("result")}, nil
	})
	if task.Name != "test-task" {
		t.Fatalf("expected name test-task, got %s", task.Name)
	}
	if task.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	// Wait for completion
	time.Sleep(100 * time.Millisecond)

	status, err := m.Status(task.ID)
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if status != StatusCompleted {
		t.Fatalf("expected Completed, got %s", status)
	}
}

func TestCreateWithError(t *testing.T) {
	m := NewManager()
	defer m.Close()

	task := m.Create("error-task", func(ctx context.Context) (Result, error) {
		return Result{Data: []byte("error data")}, errTest
	})

	time.Sleep(100 * time.Millisecond)

	status, _ := m.Status(task.ID)
	if status != StatusFailed {
		t.Fatalf("expected Failed, got %s", status)
	}

	result, err := m.GetResult(task.ID)
	if err != nil {
		t.Fatalf("GetResult error: %v", err)
	}
	if result.Err != errTest {
		t.Fatalf("expected errTest, got: %v", result.Err)
	}
}

func TestCancelRunningTask(t *testing.T) {
	m := NewManager()
	defer m.Close()

	task := m.Create("cancel-task", func(ctx context.Context) (Result, error) {
		<-ctx.Done()
		return Result{}, ctx.Err()
	})

	time.Sleep(50 * time.Millisecond)

	err := m.Cancel(task.ID)
	if err != nil {
		t.Fatalf("Cancel error: %v", err)
	}

	status, _ := m.Status(task.ID)
	if status != StatusCancelled {
		t.Fatalf("expected Cancelled, got %s", status)
	}
}

func TestCancelAlreadyCompleted(t *testing.T) {
	m := NewManager()
	defer m.Close()

	task := m.Create("done-task", func(ctx context.Context) (Result, error) {
		return Result{}, nil
	})

	time.Sleep(100 * time.Millisecond)

	// First cancel should work (task already completed, but status check)
	_ = m.Cancel(task.ID)

	// Second cancel should fail
	err := m.Cancel(task.ID)
	if err == nil {
		t.Fatal("expected error for already done task")
	}
}

func TestCancelNotFound(t *testing.T) {
	m := NewManager()
	defer m.Close()

	err := m.Cancel("nonexistent")
	if err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestStatusNotFound(t *testing.T) {
	m := NewManager()
	defer m.Close()

	_, err := m.Status("nonexistent")
	if err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestGetResultNotFound(t *testing.T) {
	m := NewManager()
	defer m.Close()

	_, err := m.GetResult("nonexistent")
	if err != ErrTaskNotFound {
		t.Fatalf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestWaitForCompletion(t *testing.T) {
	m := NewManager()
	defer m.Close()

	task := m.Create("wait-task", func(ctx context.Context) (Result, error) {
		time.Sleep(50 * time.Millisecond)
		return Result{Data: []byte("done")}, nil
	})

	err := m.WaitFor(task.ID, context.Background())
	if err != nil {
		t.Fatalf("WaitFor error: %v", err)
	}
}

func TestWaitForTimeout(t *testing.T) {
	m := NewManager()
	defer m.Close()

	task := m.Create("slow-task", func(ctx context.Context) (Result, error) {
		time.Sleep(5 * time.Second)
		return Result{}, nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := m.WaitFor(task.ID, ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunningCount(t *testing.T) {
	m := NewManager()
	defer m.Close()

	if m.RunningCount() != 0 {
		t.Fatal("expected 0 running tasks")
	}

	_ = m.Create("running-task", func(ctx context.Context) (Result, error) {
		time.Sleep(200 * time.Millisecond)
		return Result{}, nil
	})

	time.Sleep(50 * time.Millisecond)

	if m.RunningCount() != 1 {
		t.Fatalf("expected 1 running task, got %d", m.RunningCount())
	}

	time.Sleep(200 * time.Millisecond)

	if m.RunningCount() != 0 {
		t.Fatalf("expected 0 running tasks, got %d", m.RunningCount())
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusPending, "pending"},
		{StatusRunning, "running"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Fatalf("expected %s, got %s", tt.expected, tt.status.String())
		}
	}
}

var errTest = newTestError("test error")

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

func newTestError(msg string) error {
	return &testError{msg: msg}
}
