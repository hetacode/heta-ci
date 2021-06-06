package enums

type BuildResultStatus string

const (
	BuildStatusNone              BuildResultStatus = "none"
	BuildStatusRunning           BuildResultStatus = "running"
	BuildStatusFinishWithSucces  BuildResultStatus = "finish_with_success"
	BuildStatusFinishWithFailure BuildResultStatus = "finish_with_failure"
)
