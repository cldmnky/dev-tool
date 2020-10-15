package rancher

const (
	notFoundErr = "notFound"
)

type rancherError struct {
	err       string //error description
	errorType string //error type which caused the error
}

func (e *rancherError) Error() string {
	return e.err
}

func (e *rancherError) notFound() bool {
	return e.errorType == notFoundErr
}
