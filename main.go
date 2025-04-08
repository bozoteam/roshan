package main

import (
	routes "github.com/bozoteam/roshan/src"
	"github.com/bozoteam/roshan/src/helpers"
)

func main() {
	helpers.LoadDotEnv()

	routes.RunServer()
}
