package deploy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ReleaseRecord struct {
	Project           string        `json:"project"`
	Environment       string        `json:"environment"`
	ReleaseID         string        `json:"release_id"`
	PreviousReleaseID string        `json:"previous_release_id,omitempty"`
	Git               GitMetadata   `json:"git"`
	ImageTags         []string      `json:"image_tags"`
	ImageTar          string        `json:"image_tar"`
	Images            []ImageRecord `json:"images,omitempty"`
	BundlePath        string        `json:"bundle_path"`
	RemoteBase        string        `json:"remote_base"`
	Hosts             []HostRecord  `json:"hosts"`
	Migration         StepRecord    `json:"migration"`
	Apply             StepRecord    `json:"apply"`
	Rollback          StepRecord    `json:"rollback"`
	Checks            []CheckRecord `json:"checks,omitempty"`
	Status            string        `json:"status"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type ImageRecord struct {
	ID   string   `json:"id"`
	Tar  string   `json:"tar"`
	Tags []string `json:"tags"`
}

type HostRecord struct {
	ID        string   `json:"id"`
	SSH       string   `json:"ssh"`
	Services  []string `json:"services"`
	RemoteDir string   `json:"remote_dir"`
	Status    string   `json:"status"`
}

type StepRecord struct {
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
	At          time.Time `json:"at,omitempty"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	FinishedAt  time.Time `json:"finished_at,omitempty"`
	Command     []string  `json:"command,omitempty"`
	Image       string    `json:"image,omitempty"`
	EnvSource   string    `json:"env_source,omitempty"`
	RestorePlan string    `json:"restore_plan,omitempty"`
}

type CheckRecord struct {
	Group           string    `json:"group"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	Target          string    `json:"target"`
	ExpectedStatus  int       `json:"expected_status,omitempty"`
	ActualStatus    int       `json:"actual_status,omitempty"`
	FollowRedirects bool      `json:"follow_redirects,omitempty"`
	EnvSource       string    `json:"env_source,omitempty"`
	Status          string    `json:"status"`
	Error           string    `json:"error,omitempty"`
	StartedAt       time.Time `json:"started_at"`
	FinishedAt      time.Time `json:"finished_at"`
	DurationMS      int64     `json:"duration_ms"`
}

type State struct {
	Project           string    `json:"project"`
	Environment       string    `json:"environment"`
	CurrentReleaseID  string    `json:"current_release_id,omitempty"`
	PreviousReleaseID string    `json:"previous_release_id,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func LoadState(root string) (State, error) {
	path := statePath(root)
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("read state: %w", err)
	}
	var state State
	if err := json.Unmarshal(body, &state); err != nil {
		return State{}, fmt.Errorf("parse state: %w", err)
	}
	return state, nil
}

func SaveState(root string, state State) error {
	state.UpdatedAt = time.Now().UTC()
	if err := os.MkdirAll(filepath.Dir(statePath(root)), 0755); err != nil {
		return fmt.Errorf("create .deploy: %w", err)
	}
	return writeJSON(statePath(root), state)
}

func LoadRelease(root string, releaseID string) (ReleaseRecord, error) {
	body, err := os.ReadFile(releaseMetadataPath(root, releaseID))
	if err != nil {
		return ReleaseRecord{}, fmt.Errorf("read release %s: %w", releaseID, err)
	}
	var record ReleaseRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return ReleaseRecord{}, fmt.Errorf("parse release %s: %w", releaseID, err)
	}
	return record, nil
}

func SaveRelease(root string, record ReleaseRecord) error {
	record.UpdatedAt = time.Now().UTC()
	if err := os.MkdirAll(filepath.Dir(releaseMetadataPath(root, record.ReleaseID)), 0755); err != nil {
		return fmt.Errorf("create release metadata dir: %w", err)
	}
	return writeJSON(releaseMetadataPath(root, record.ReleaseID), record)
}

func statePath(root string) string {
	return filepath.Join(root, ".deploy", "state.json")
}

func releaseMetadataPath(root string, releaseID string) string {
	return filepath.Join(root, ".deploy", "releases", releaseID, "metadata.json")
}

func writeJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
