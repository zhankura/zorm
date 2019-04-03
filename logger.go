package zorm


type logger interface {
	Print(v ...interface{})
}

type LogWriter interface {
	Println(v ...interface{})
}

type Logger struct {
	LogWriter
}

func (logger Logger) Print(values ...interface{}) {
	logger.Println("")
}