package structs

type Pipeline struct {
	Name string
	Jobs []Job
}

type Job struct {
	ID          string     `json:"id"` // unique name on jobs level - no whitespaces, no special chars
	DisplayName string     `json:"name"`
	Runner      string     `json:"runner"`
	Tasks       []Task     `json:"tasks"`
	Conditons   []Conditon `json:"conditions"`
}

type Task struct {
	ID          string     `json:"id"` // unique name on tasks level - no whitespaces, no special chars
	DisplayName string     `json:"name"`
	Command     []string   `json:"command"` // one or more
	Conditons   []Conditon `json:"conditions"`
}

type Conditon struct {
	// when run
	Type ConditionType `json:"type"`
	// which job/task should trigger this condition "owner"
	On string `json:"on"`
}

type ConditionType string

const (
	OnFailure ConditionType = "on_failure"
	OnSuccess               = "on_success"
)
