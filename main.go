package main

import (
	"github.com/polygon-fast-lane/calldata-generator/generator"
)

func main() {
	h := generator.NewHarness()
	h.Run()
	h.TestSwapIntent()
}
