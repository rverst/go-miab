package main

import "github.com/rverst/go-miab/command"

const Version = "1.0.0-beta1"

func main() {
	command.Execute(Version)
}
