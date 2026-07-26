package log

import (
	"fmt"
	"os"
)

var quiet bool

func SetQuiet(q bool) { quiet = q }

func Logf(format string, a ...any) {
	if !quiet {
		fmt.Printf(format+"\n", a...)
	}
}

func Println(a ...any) {
	if !quiet {
		fmt.Println(a...)
	}
}

func Fatalln(a ...any) {
	if !quiet {
		fmt.Println(a...)
	}

	os.Exit(1)
}
