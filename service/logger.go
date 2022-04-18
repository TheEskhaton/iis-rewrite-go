package service

import "fmt"

type Logger interface {
	LogF(format string, a ...interface{})
	LogLn(string)
}

type stdLogger struct {
}

func (l *stdLogger) LogF(format string, a ...interface{}) {
	logF(format, a...)
}

func (l *stdLogger) LogLn(output string) {
	logLn(output)
}

type nilLogger struct {
}

func (l *nilLogger) LogF(format string, a ...interface{}) {

}

func (l *nilLogger) LogLn(output string) {

}

func logLn(output string) {
	fmt.Println(output)
}

func logF(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func NewLogger(silent bool) Logger {
	if silent {
		return &nilLogger{}
	}

	return &stdLogger{}
}
