package log

import "fmt"

type Logger struct {
	Level uint8
}

func MustGetLevel(name string) uint8 {
	switch name {
	case "Error":
		return 1
	case "Warning":
		return 2
	case "Attention":
		return 3
	case "Notice":
		return 4
	case "Mark":
		return 5
	case "News":
		return 6
	case "Info":
		return 7
	case "Noise":
		return 8
	case "Debug":
		return 9
	case "Nothing":
		return 10
	case "Ignore":
		return 200
	default:
		panic("Unsupported Log Level: " + name)
	}
}

/*
0 Log.Fatal()
1 Log.Error()
2 Log.Warning()
3 Log.Attention()
4 Log.Notice()
5 Log.Mark()
6 Log.News()
7 Log.Info()
8 Log.Noise()
9 Log.Debug()
10 Log.Nothing()
200 Log.Ignore()
*/

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

func (lg *Logger) Fatal(stuff ...interface{}) {
	fmt.Println(stuff...)
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

func (lg *Logger) Notice(stuff ...interface{}) {
	if lg.Level >= 4 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Mark(stuff ...interface{}) {
	if lg.Level >= 5 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) News(stuff ...interface{}) {
	if lg.Level >= 6 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Info(stuff ...interface{}) {
	if lg.Level >= 7 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Noise(stuff ...interface{}) {
	if lg.Level >= 8 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Debug(stuff ...interface{}) {
	if lg.Level >= 9 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Nothing(stuff ...interface{}) {
	if lg.Level >= 10 {
		fmt.Println(stuff...)
	}
}

func (lg *Logger) Ignore(stuff ...interface{}) {
	if lg.Level >= 200 {
		fmt.Println(stuff...)
	}
}
