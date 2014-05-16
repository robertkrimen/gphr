package main

import (
	"fmt"
	"os"
)

var (
	lg  = _log{}
	log = lg.log
)

type _log struct {
}

func (lg _log) log(format string, arguments ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", arguments...)
}

func (lg _log) err(format string, arguments ...interface{}) {
	fmt.Fprintf(os.Stderr, "gphr: "+format+"\n", arguments...)
}

func (lg _log) error(format string, arguments ...interface{}) error {
	return fmt.Errorf(format, arguments...)
}

func (lg _log) dbg(format string, arguments ...interface{}) {
	if !*flags.main.debug {
		return
	}
	fmt.Fprintf(os.Stderr, "gphr: "+format+"\n", arguments...)
}

func debug(format string, arguments ...interface{}) {
	if !*flags.main.debug {
		return
	}
	fmt.Fprintf(os.Stderr, format, arguments...)
}
