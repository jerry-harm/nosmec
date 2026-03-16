/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/jerry-harm/nosmec/cmd"
	"github.com/jerry-harm/nosmec/config"
)

func main() {
	cmd.Execute()
	config.SaveRelays()
}
