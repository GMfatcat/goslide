package main

import "github.com/user/goslide/internal/cli"

var version = "dev"

func main() {
	cli.Execute(version)
}
