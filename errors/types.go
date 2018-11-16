package errors

import (
	"fmt"
)

type baseErr struct {
	code int
	msg  string
}

func (err baseErr) Code() int {
	return err.code
}

func (err baseErr) Message() string {
	return err.msg
}

func (err baseErr) Error() string {
	return fmt.Sprintf("%d: %s", err.code, err.msg)
}

func (err baseErr) String() string {
	return err.Error()
}

type requestError struct {
	baseErr
	statusCode int
}

func (err requestError) StatusCode() int {
	return err.statusCode
}

func (err requestError) Error() string {
	return fmt.Sprintf("%s with status code : %d", err.baseErr.Error(), err.statusCode)
}

func (err requestError) String() string {
	return err.Error()
}
