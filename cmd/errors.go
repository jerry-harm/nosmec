package cmd

import (
	"fmt"
	"os"

	"github.com/jerry-harm/nosmec/logger"
)

type CommandError struct {
	Message string
	Err     error
}

func (e *CommandError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

func newError(msg string, err error) *CommandError {
	return &CommandError{Message: msg, Err: err}
}

func fatal(msg string, err error) {
	if err != nil {
		logger.Error(msg, "error", err.Error())
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", msg, err)
	} else {
		logger.Error(msg)
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
	os.Exit(1)
}

func fatalf(msg string, args ...interface{}) {
	logger.Error(fmt.Sprintf(msg, args...))
	fmt.Fprintf(os.Stderr, "Error: "+msg+"\n", args...)
	os.Exit(1)
}

func handleError(err error) {
	if err == nil {
		return
	}
	if cmdErr, ok := err.(*CommandError); ok {
		fatal(cmdErr.Message, cmdErr.Err)
	}
	fatal("operation failed", err)
}