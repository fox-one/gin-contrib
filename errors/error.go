package errors

type Error interface {
	error

	Code() int
	Message() string
}

type RequestError interface {
	Error

	StatusCode() int
}

func New(code int, message string, status ...int) Error {
	err := baseErr{code, message}
	if len(status) > 0 {
		return &requestError{baseErr: err, statusCode: status[0]}
	}

	return &err
}
