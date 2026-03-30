package coordination

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

type WaveExecutor struct {
	coordinator *Coordinator
}

func NewWaveExecutor(coord *Coordinator) *WaveExecutor {
	return &WaveExecutor{coordinator: coord}
}

func (we *WaveExecutor) CalculateWaves(tasks []*Task) [][]*Task {
	remaining := make(map[string]*Task)
	for _, t := range tasks {
		remaining[t.ID] = t
	}

	completed := make(map[string]bool)
	var waves [][]*Task

	for len(remaining) > 0 {
		var wave []*Task
		var blocked []*Task

		for _, task := range remaining {
			allDepsMet := true
			for _, depID := range task.BlockedBy {
				if !completed[depID] {
					allDepsMet = false
					break
				}
			}
			if allDepsMet {
				wave = append(wave, task)
			} else {
				blocked = append(blocked, task)
			}
		}

		if len(wave) == 0 {
			log.Error().Int("blocked", len(blocked)).Msg("deadlock detected in wave execution")
			break
		}

		for _, t := range wave {
			t.Wave = len(waves) + 1
			we.coordinator.UpdateTaskStatus(t.ID, StatusPending, nil, "")
		}

		waves = append(waves, wave)

		for _, t := range wave {
			completed[t.ID] = true
			delete(remaining, t.ID)
		}
	}

	return waves
}

func (we *WaveExecutor) ExecuteWaves(ctx context.Context, waves [][]*Task, executeFn func(context.Context, *Task) (interface{}, error)) error {
	for waveNum, wave := range waves {
		log.Info().Int("wave", waveNum+1).Int("tasks", len(wave)).Msg("starting wave execution")

		var wg sync.WaitGroup
		errCh := make(chan error, len(wave))
		resultCh := make(chan struct {
			taskID string
			output interface{}
		}, len(wave))

		for _, task := range wave {
			wg.Add(1)
			go func(t *Task) {
				defer wg.Done()

				we.coordinator.UpdateTaskStatus(t.ID, StatusInProgress, nil, "")

				output, err := executeFn(ctx, t)
				if err != nil {
					errCh <- fmt.Errorf("task %s failed: %w", t.ID, err)
					we.coordinator.UpdateTaskStatus(t.ID, StatusFailed, nil, err.Error())
					return
				}

				resultCh <- struct {
					taskID string
					output interface{}
				}{taskID: t.ID, output: output}
				we.coordinator.CompleteTask(t.ID, output)
			}(task)
		}

		wg.Wait()
		close(errCh)
		close(resultCh)

		for err := range errCh {
			return err
		}

		for result := range resultCh {
			log.Info().Str("task", result.taskID).Msg("task completed in wave")
		}
	}

	return nil
}
