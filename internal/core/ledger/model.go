package ledger

import "time"

type Step struct {
	StepID     string    `json:"step_id"`
	Kind       StepKind  `json:"kind"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	DurationMs int64     `json:"duration_ms,omitempty"`

	Cmd          []string          `json:"cmd,omitempty"`
	Cwd          string            `json:"cwd,omitempty"`
	Env          map[string]string `json:"env,omitempty"`
	ExitCode     *int              `json:"exit_code,omitempty"`
	DiffStat     *DiffStat         `json:"diff_stat,omitempty"`
	Artifacts    *Artifacts        `json:"artifacts,omitempty"`
	PolicyEvents []PolicyEvent     `json:"policy_events,omitempty"`

	Mode          string `json:"mode,omitempty"`
	CommitSHA     string `json:"commit_sha,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
	TargetBranch  string `json:"target_branch,omitempty"`

	Target     string `json:"target,omitempty"`
	TargetStep string `json:"target_step,omitempty"`
	Hard       *bool  `json:"hard,omitempty"`
}

type StepKind string

const (
	StepKindRun      StepKind = "run"
	StepKindApply    StepKind = "apply"
	StepKindRollback StepKind = "rollback"
)

type DiffStat struct {
	Files     int      `json:"files"`
	Additions int      `json:"additions"`
	Deletions int      `json:"deletions"`
	FileList  []string `json:"file_list,omitempty"`
}

type Artifacts struct {
	Patch  string `json:"patch,omitempty"`
	Output string `json:"output,omitempty"`
}

type PolicyEvent struct {
	Rule    string `json:"rule"`
	Action  string `json:"action"`
	Matched string `json:"matched"`
}
