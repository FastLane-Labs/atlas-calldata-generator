package generator

import (
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func (h *Harness) GenerateDAppCalldata(dapp string, funcName string, args string) (string, error) {
	dappAddr, ok := h.AddressMap[dapp]
	if !ok {
		return "", errors.New("err - DApp Address not found A")
	}

	contract, ok := h.DAppMap[dappAddr]
	if !ok {
		return "", errors.New("err - DApp Contract not found B")
	}

	result, err := contract.EncodeTxData(funcName, args)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(result), nil
}
