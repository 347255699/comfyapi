/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/347255699/comfyapi/cmd"
)

func main() {
	if err := cmd.Commands().Execute(); err != nil {
		panic(err)
	}
}
