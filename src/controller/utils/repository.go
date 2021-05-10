package utils

// Repository represent a list of all git repositories in order to listening of code changes
// and if any changes occur should run pipeline which definition file should exists inside repo in specifig directory
type Repository struct {
	Url           string
	DefaultBranch string
}
