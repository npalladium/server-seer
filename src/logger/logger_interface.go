package logger

var Logger LoggerInterface

type LoggerInterface interface {
	Log(string)
}
