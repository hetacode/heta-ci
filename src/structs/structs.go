package structs

type Pipeline struct {
	RepositoryID string  `json:"repository_id"`
	Name         string  `yaml:"name"`
	Jobs         []Job   `yaml:"jobs"`
	RunOn        []RunOn `ymal:"run_on"` // determine conditions when build should run
}

type Job struct {
	ID          string     `yaml:"id" json:"id"` // unique name on jobs level - no whitespaces, no special chars
	DisplayName string     `yaml:"display_name" json:"name"`
	Runner      string     `yaml:"runner" json:"runner"`
	Tasks       []Task     `yaml:"tasks" json:"tasks"`
	Conditons   []Conditon `yaml:"conditions" json:"conditions"`
}

type Task struct {
	ID          string     `yaml:"id" json:"id"` // unique name on tasks level - no whitespaces, no special chars
	DisplayName string     `yaml:"display_name" json:"name"`
	Command     string     `yaml:"command" json:"command"` // one or more
	Conditons   []Conditon `yaml:"conditions" json:"conditions"`
}

type Conditon struct {
	// when run
	Type ConditionType `yaml:"type" json:"type"`
	// which job/task should trigger this condition "owner"
	On string `yaml:"on" json:"on"`
}

type RunOn struct {
	Type RunOnType `yaml:"type"`
	On   string    `yaml:"on"`
}

type RunOnType string

const (
	// if any branch condition is fit
	RunOnBranch RunOnType = "branch"
)

type ConditionType string

const (
	OnFailure ConditionType = "on_failure"
	OnSuccess               = "on_success"
)
