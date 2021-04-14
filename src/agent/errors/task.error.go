package errors

type ContainerError struct {
	ErrorCode int
	Message   string
}

func (e *ContainerError) Error() string {
	return e.Message
}
