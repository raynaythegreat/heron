package learning

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Instinct struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Trigger     string    `json:"trigger"`
	Solution    string    `json:"solution"`
	Confidence  float64   `json:"confidence"`
	TimesUsed   int       `json:"times_used"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Verified    bool      `json:"verified"`
}

type InstinctStore struct {
	instinctsPath string
}

func NewInstinctStore() *InstinctStore {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".heron", "workspace", "instincts")
	os.MkdirAll(path, 0755)
	return &InstinctStore{instinctsPath: path}
}

func (s *InstinctStore) ExtractInstinct(name, description, trigger, solution string, tags []string) (*Instinct, error) {
	instinct := &Instinct{
		ID:          fmt.Sprintf("instinct_%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Trigger:     trigger,
		Solution:    solution,
		Confidence:  0.5,
		TimesUsed:   0,
		CreatedAt:   time.Now(),
		Tags:        tags,
		Verified:    false,
	}

	data, _ := json.MarshalIndent(instinct, "", "  ")
	path := filepath.Join(s.instinctsPath, instinct.ID+".json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, err
	}

	return instinct, nil
}

func (s *InstinctStore) SearchInstincts(query string) ([]*Instinct, error) {
	entries, err := os.ReadDir(s.instinctsPath)
	if err != nil {
		return nil, err
	}

	var results []*Instinct
	queryLower := strings.ToLower(query)

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.instinctsPath, entry.Name()))
		if err != nil {
			continue
		}
		var instinct Instinct
		if err := json.Unmarshal(data, &instinct); err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(instinct.Trigger), queryLower) ||
			strings.Contains(strings.ToLower(instinct.Description), queryLower) ||
			strings.Contains(strings.ToLower(instinct.Name), queryLower) {
			results = append(results, &instinct)
		}
	}

	return results, nil
}

func (s *InstinctStore) RecordUsage(instinctID string) error {
	data, err := os.ReadFile(filepath.Join(s.instinctsPath, instinctID+".json"))
	if err != nil {
		return err
	}
	var instinct Instinct
	if err := json.Unmarshal(data, &instinct); err != nil {
		return err
	}

	instinct.TimesUsed++
	instinct.LastUsedAt = time.Now()
	if instinct.TimesUsed >= 3 && instinct.Verified {
		instinct.Confidence = min(0.99, instinct.Confidence+0.1)
	} else if instinct.TimesUsed >= 3 {
		instinct.Verified = true
		instinct.Confidence = 0.8
	}

	updated, _ := json.MarshalIndent(instinct, "", "  ")
	return os.WriteFile(filepath.Join(s.instinctsPath, instinctID+".json"), updated, 0644)
}

func (s *InstinctStore) ExportToSkill(instinctID string) (string, error) {
	data, err := os.ReadFile(filepath.Join(s.instinctsPath, instinctID+".json"))
	if err != nil {
		return "", err
	}
	var instinct Instinct
	if err := json.Unmarshal(data, &instinct); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", instinct.Name))
	sb.WriteString(fmt.Sprintf("description: %s\n", instinct.Description))
	sb.WriteString(fmt.Sprintf("trigger: %s\n", instinct.Trigger))
	sb.WriteString(fmt.Sprintf("confidence: %.2f\n", instinct.Confidence))
	if len(instinct.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("tags: %s\n", strings.Join(instinct.Tags, ", ")))
	}
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("# %s\n\n", instinct.Name))
	sb.WriteString(fmt.Sprintf("## When to Use\n%s\n\n", instinct.Trigger))
	sb.WriteString(fmt.Sprintf("## Solution\n%s\n", instinct.Solution))

	return sb.String(), nil
}
