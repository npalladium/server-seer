package logger

import (
	"fmt"
)

type ScreenLogger struct {
}

func (self ScreenLogger) Log(content string) {
	fmt.Println(content)
}
