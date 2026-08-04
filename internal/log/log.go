package log

import (
	"fmt"
	"os"
)

type level uint8

const (
	DEBUG = level(iota)
	INFO
	WARN
	ERROR
	CRITICAL
)

var (
	minLevel = INFO
)

func SetQuiet(q bool)   { minLevel = ERROR }
func SetVerbose(v bool) { minLevel = DEBUG }

func Logf(lv level, format string, a ...any) {
	if lv >= minLevel {
		fmt.Printf(format, a...)
	}
}

func Println(lv level, a ...any) {
	if lv >= minLevel {
		fmt.Println(a...)
	}
}

func Fatalln(a ...any) {
	fmt.Println(a...)
	os.Exit(1)
}

func Fatalf(format string, a ...any) {
	fmt.Printf(format, a...)
	os.Exit(1)
}
