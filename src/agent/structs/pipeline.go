package structs

type Pipeline struct {
	Name string
	Jobs []Job
}

type Job struct {
	Name   string
	Runner string
	Tasks  []Task
}

type Task struct {
	Name    string
	Command []string // one or more
}
