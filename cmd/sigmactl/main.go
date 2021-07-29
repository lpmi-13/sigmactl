package main

import (
	"log"

	"github.com/lpmi-13/sigmactl/commands"
)

func main() {
	log.SetPrefix("sigmactl: ")
	commands.Execute()
}
