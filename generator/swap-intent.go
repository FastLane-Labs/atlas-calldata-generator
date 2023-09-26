package generator

import "fmt"

const (
	swapDAppFunc string = "swap"
	swapDAppName string = "SwapIntentController"
)

func (h *Harness) printSwapArgs() error {
	err := h.printArgs(swapDAppName, swapDAppFunc)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (h *Harness) TestSwapIntent() error {
	fmt.Println("TestSwapIntent()")

	err := h.printArgs(swapDAppName, swapDAppFunc)
	if err != nil {
		fmt.Println(err)
		return err
	}

	h.printSwapArgs()

	calldata, err := h.GenerateDAppCalldata(swapDAppName, swapDAppFunc, SwapIntentExample)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("DApp calldata: " + calldata)

	return nil
}

var SwapIntentExample = string(`
	[
		{
			"tokenUserBuys": "0x6B175474E89094C44Da98b954EedeAC495271d0F",
			"amountUserBuys": 1500000000000000000000,
			"tokenUserSells": "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
			"amountUserSells": 1000000000000000000,
			"auctionBaseCurrency": "0x0000000000000000000000000000000000000000",
			"solverMustReimburseGas": false,
			"conditions": []
		}
	]
`)
