package cmd

import (
	"github.com/spf13/cobra"
)

type CommandRegistrar func(*cobra.Command) error

var commandRegistrar CommandRegistrar

func RegisterCommands(fn CommandRegistrar) {
	commandRegistrar = fn
}

func initCommands() {
	if commandRegistrar == nil {
		registerDefaultCommands()
	} else {
		commandRegistrar(rootCmd)
	}
}

func registerDefaultCommands() {
	registerNoteCommands()
	registerConfigCommands()
	registerProfileCommands()
	registerDMCommands()
	registerCommunityCommands()
	registerEventCommands()
	registerSearchCommands()
	registerGossipCommands()
	registerRelayCommands()
}

type commandGroup struct {
	name        string
	description string
	commands    []*cobra.Command
}

var commandGroups []commandGroup

func RegisterCommandGroup(name, description string, cmds ...*cobra.Command) {
	commandGroups = append(commandGroups, commandGroup{
		name:        name,
		description: description,
		commands:    cmds,
	})
	for _, cmd := range cmds {
		rootCmd.AddCommand(cmd)
	}
}

func GetCommandGroups() []commandGroup {
	return commandGroups
}
