package log

import "fmt"

type Logger struct {
	Level uint8
}

func MustGetLevel(name string) uint8 {
	switch name {
	// error
	case "Error":
		return 1
	case "Warning":
		return 2
	case "Attention":
		return 3
	// data
	case "Note":
		return 1
	case "News":
		return 2
	case "Info":
		return 3
	// debug
	case "Debug":
		return 4
	case "Noise":
		return 5

	default:
		panic("Unsupported Log Level: " + name)
	}
}

func New() Logger {
	return Logger{
		Level: 255,
	}
}

func NewWithLevel(level uint8) Logger {
	return Logger{
		Level: level,
	}
}

func (lg *Logger) Error(stuff ...interface{}) {
	if lg.Level >= 1 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Warning(stuff ...interface{}) {
	if lg.Level >= 2 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Attention(stuff ...interface{}) {
	if lg.Level >= 3 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Note(stuff ...interface{}) {
	if lg.Level >= 1 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) News(stuff ...interface{}) {
	if lg.Level >= 2 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Info(stuff ...interface{}) {
	if lg.Level >= 3 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Debug(stuff ...interface{}) {
	if lg.Level >= 4 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Noise(stuff ...interface{}) {
	if lg.Level >= 5 {
		fmt.Println(stuff...)
	}
}
