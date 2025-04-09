package main

import (
	"github.com/bozoteam/roshan/internal/adapter"
	"github.com/bozoteam/roshan/internal/helpers"
)

func main() {
	helpers.LoadDotEnv()

	adapter.RunServer()
}
