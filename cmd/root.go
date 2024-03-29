/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cfy",
		Short: "cfy is a comfyui cli",
	}

	cmd.AddCommand(QueuePrompt())
	cmd.AddCommand(Watch())
	return cmd
}
