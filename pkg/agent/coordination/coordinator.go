package coordination

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
	StatusBlocked    TaskStatus = "blocked"
)

type Task struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	AgentID      string      `json:"agent_id"`
	Status       TaskStatus  `json:"status"`
	Priority     int         `json:"priority"`
	Wave         int         `json:"wave"`
	BlockedBy    []string    `json:"blocked_by,omitempty"`
	Assignee     string      `json:"assignee,omitempty"`
	Input        interface{} `json:"input,omitempty"`
	Output       interface{} `json:"output,omitempty"`
	WorktreePath string      `json:"worktree_path,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	StartedAt    time.Time   `json:"started_at,omitempty"`
	CompletedAt  time.Time   `json:"completed_at,omitempty"`
	Error        string      `json:"error,omitempty"`
}

type AgentMessage struct {
	ID        string    `json:"id"`
	FromAgent string    `json:"from_agent"`
	ToAgent   string    `json:"to_agent"`
	TaskID    string    `json:"task_id,omitempty"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Coordinator struct {
	basePath string
	mu       sync.RWMutex
	agents   map[string]string
}

func NewCoordinator(workspacePath string) *Coordinator {
	home, _ := os.UserHomeDir()
	basePath := filepath.Join(home, ".heron", "coordination")
	os.MkdirAll(filepath.Join(basePath, "inbox"), 0755)
	os.MkdirAll(filepath.Join(basePath, "outbox"), 0755)
	os.MkdirAll(filepath.Join(basePath, "locks"), 0755)
	os.MkdirAll(filepath.Join(basePath, "tasks"), 0755)
	os.MkdirAll(filepath.Join(basePath, "completed"), 0755)

	return &Coordinator{
		basePath: basePath,
		agents:   make(map[string]string),
	}
}

func (c *Coordinator) CreateTask(task *Task) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	task.CreatedAt = time.Now()
	if task.Status == "" {
		task.Status = StatusPending
	}

	data, _ := json.MarshalIndent(task, "", "  ")
	tmpPath := filepath.Join(c.basePath, "tasks", task.ID+".tmp")
	finalPath := filepath.Join(c.basePath, "tasks", task.ID+".json")

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, finalPath)
}

func (c *Coordinator) GetTask(taskID string) (*Task, error) {
	data, err := os.ReadFile(filepath.Join(c.basePath, "tasks", taskID+".json"))
	if err != nil {
		return nil, err
	}
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (c *Coordinator) UpdateTaskStatus(taskID string, status TaskStatus, output interface{}, errMsg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	task, err := c.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = status
	if status == StatusInProgress {
		task.StartedAt = time.Now()
	}
	if status == StatusCompleted || status == StatusFailed {
		task.CompletedAt = time.Now()
	}
	if output != nil {
		task.Output = output
	}
	if errMsg != "" {
		task.Error = errMsg
	}

	data, _ := json.MarshalIndent(task, "", "  ")
	return os.WriteFile(filepath.Join(c.basePath, "tasks", taskID+".json"), data, 0644)
}

func (c *Coordinator) ListPendingTasks() ([]*Task, error) {
	entries, err := os.ReadDir(filepath.Join(c.basePath, "tasks"))
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(c.basePath, "tasks", entry.Name()))
		if err != nil {
			continue
		}
		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			continue
		}
		if task.Status == StatusPending || task.Status == StatusBlocked {
			tasks = append(tasks, &task)
		}
	}
	return tasks, nil
}

func (c *Coordinator) SendMessage(msg *AgentMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	msg.Timestamp = time.Now()

	inboxPath := filepath.Join(c.basePath, "inbox", msg.ToAgent)
	os.MkdirAll(inboxPath, 0755)

	data, _ := json.MarshalIndent(msg, "", "  ")
	tmpPath := filepath.Join(inboxPath, msg.ID+".tmp")
	finalPath := filepath.Join(inboxPath, msg.ID+".json")

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, finalPath)
}

func (c *Coordinator) ReceiveMessages(agentID string) ([]*AgentMessage, error) {
	inboxPath := filepath.Join(c.basePath, "inbox", agentID)
	entries, err := os.ReadDir(inboxPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var messages []*AgentMessage
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(inboxPath, entry.Name()))
		if err != nil {
			continue
		}
		var msg AgentMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (c *Coordinator) AcquireLock(resourceID, agentID string) (bool, error) {
	lockPath := filepath.Join(c.basePath, "locks", resourceID+".lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return false, nil
		}
		return false, err
	}
	f.WriteString(agentID + "\n")
	f.Close()
	return true, nil
}

func (c *Coordinator) ReleaseLock(resourceID, agentID string) error {
	lockPath := filepath.Join(c.basePath, "locks", resourceID+".lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return err
	}
	if string(data) == agentID+"\n" {
		return os.Remove(lockPath)
	}
	return fmt.Errorf("lock held by different agent")
}

func (c *Coordinator) CompleteTask(taskID string, output interface{}) error {
	task, err := c.GetTask(taskID)
	if err != nil {
		return err
	}

	task.Status = StatusCompleted
	task.CompletedAt = time.Now()
	task.Output = output

	data, _ := json.MarshalIndent(task, "", "  ")
	completedPath := filepath.Join(c.basePath, "completed", taskID+".json")
	if err := os.WriteFile(completedPath, data, 0644); err != nil {
		return err
	}

	taskPath := filepath.Join(c.basePath, "tasks", taskID+".json")
	return os.Remove(taskPath)
}

func (c *Coordinator) GetBasePath() string {
	return c.basePath
}
