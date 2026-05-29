package main

import (
	"os"

	"github.com/jerry-harm/nosmec/cmd"
	"github.com/jerry-harm/nosmec/gui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "cli" {
		cmd.Execute()
		return
	}
	gui.Run()
}