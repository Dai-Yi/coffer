package main

import (
	"coffer/cmd"
	"log"
	"time"
)

func init() {
	timelocal := time.FixedZone("CST", 3600*8)
	time.Local = timelocal
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}
func main() {
	cmd.CMDControl()
}
