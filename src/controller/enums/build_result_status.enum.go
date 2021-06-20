package enums

type BuildStatus string

const (
	BuildStatusNone              BuildStatus = "none"
	BuildStatusRunning           BuildStatus = "running"
	BuildStatusFinishWithSucces  BuildStatus = "finish_with_success"
	BuildStatusFinishWithFailure BuildStatus = "finish_with_failure"
)
