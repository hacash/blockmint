package config

import "github.com/hacash/blockmint/sys/log"

////////////// Logger ////////////////

var (
	globalLogger *log.Logger = nil
)

func GetGlobalInstanceLogger() log.Logger {
	lv := log.MustGetLevel(Config.Loglevel)
	return log.Logger{
		Level: lv, //
	}
}
