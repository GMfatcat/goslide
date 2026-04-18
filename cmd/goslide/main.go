package main

import "github.com/GMfatcat/goslide/internal/cli"

var version = "dev"

func main() {
	cli.Execute(version)
}
