package zorm

import (
	"fmt"
	"os"
	"time"
)

type Logger interface {
	Print(v ...interface{})
}

type LogWriter interface {
	Println(v ...interface{})
}

type logger struct {
	LogWriter
	outFile *os.File
}

func (l logger) Print(values ...interface{}) {
	out := l.outFile
	end := time.Now()
	formatStr := fmt.Sprintf("[ZORM] %v", end.Format("2006/01/02 - 15:04:05"))
	for _, value := range values {
		result := fmt.Sprintf(" %v", value)
		formatStr += result
	}
	formatStr += "\n"
	fmt.Fprintf(out, formatStr)
}
