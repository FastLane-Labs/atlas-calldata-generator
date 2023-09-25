package main

import (
	"errors"
	"fmt"
)

func (h *Harness) loadDApp(dapp string) error {
	if dapp == "Atlas" {
		return errors.New("err - Atlas is not a DApp")
	}

	swapIntentAddr, ok := h.AddressMap[dapp]
	if !ok {
		return errors.New("err - " + dapp + " not in config")
	}
	var path string = "./abis/" + dapp + ".json"

	err := h.loadContractOS(swapIntentAddr, path)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	h := NewHarness()
	for dapp := range h.AddressMap {
		if dapp != "Atlas" {
			err := h.loadDApp(dapp)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(dapp + " DApp Loaded")
			}
		}
	}

	h.printSwapArgs()

	calldata, err := h.generateDAppCalldata(swapDAppName, swapDAppFunc, SwapIntentExample)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("DApp calldata: " + calldata)

}
