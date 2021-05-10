package main

import (
	"coffer/cmd"
	"time"
)

func init() {
	timelocal := time.FixedZone("CST", 3600*8)
	time.Local = timelocal
}
func main() {
	cmd.CMDControl()
}
