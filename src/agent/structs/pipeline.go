package structs

type Pipeline struct {
	Name string
	Jobs []Job
}

type Job struct {
	ID          string // unique name on jobs level - no whitespaces, no special chars
	DisplayName string
	Runner      string
	Tasks       []Task
	Conditons   []Conditon
}

type Task struct {
	ID          string // unique name on tasks level - no whitespaces, no special chars
	DisplayName string
	Command     []string // one or more
	Conditons   []Conditon
}

type Conditon struct {
	// when run
	Type ConditionType
	// which job/task should trigger this condition "owner"
	On string
}

type ConditionType string

const (
	OnFailure ConditionType = "on_failure"
	OnSuccess               = "on_success"
)
