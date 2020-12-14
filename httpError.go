package mux

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	env "github.com/karkael64/golang-env"
)

type HttpErrorStack struct {
	Function string
	File     string
	AbsFile  string
	Line     int
}

type HttpError struct {
	code    int
	message string
	stack   []HttpErrorStack
}

const STACK_SKIP = 3
const STACK_MAX_DEEP = 16

func NewHttpError(code int, message string) *HttpError {
	stack := GetCurrentStack(STACK_SKIP, STACK_MAX_DEEP)
	return &HttpError{code, message, stack}
}

func FromError(err error) *HttpError {
	stack := GetCurrentStack(STACK_SKIP, STACK_MAX_DEEP)
	return &HttpError{500, err.Error(), stack}
}

func (err *HttpError) GetCode() int {
	return err.code
}

func (err *HttpError) GetMessage() string {
	return err.message
}

func (err *HttpError) GetTitle() string {
	return http.StatusText(err.code)
}

func (err *HttpError) GetStack() []HttpErrorStack {
	return err.stack
}

func (err *HttpError) GetStackString() string {
	return StackToString(err.stack)
}

func (err *HttpError) Error() string {
	var isDebugging = env.HasEnv("http_debug")
	if isDebugging {
		return fmt.Sprintf(
			"%d - %s\n%s\n\nError thrown\n%s\n",
			err.GetCode(),
			err.GetTitle(),
			err.GetMessage(),
			err.GetStackString(),
		)
	} else {
		return fmt.Sprintf(
			"%d - %s\n",
			err.GetCode(),
			err.GetTitle(),
		)
	}
}

func (err *HttpError) Send(w http.ResponseWriter) {
	http.Error(w, err.Error(), err.GetCode())
}

func GetCurrentStack(skip, maxDeep int) []HttpErrorStack {
	pc := make([]uintptr, maxDeep)
	result := make([]HttpErrorStack, 0, maxDeep)

	runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc)
	dir, errdir := filepath.Abs(".")
	fullDebugging := env.GetEnv("http_debug") == "full"

	for frame, ok := frames.Next(); ok; frame, ok = frames.Next() {
		var rel = frame.File
		if errdir == nil {
			if strings.HasPrefix(frame.File, dir) {
				rel = "." + frame.File[len(dir):]
			} else {
				if !fullDebugging {
					continue
				}
			}
		}

		result = append(result, HttpErrorStack{
			frame.Function,
			rel,
			frame.File,
			frame.Line,
		})
	}
	return result
}

func StackFrameToString(frame *HttpErrorStack) string {
	return fmt.Sprintf("  at %s:%d <%s>", frame.File, frame.Line, frame.Function)
}

func StackToString(stack []HttpErrorStack) string {
	var lenStack = len(stack)
	var lines []string = make([]string, lenStack)
	for i := 0; i < lenStack; i++ {
		lines[i] = StackFrameToString(&stack[i])
	}
	return strings.Join(lines, "\n")
}
