package main

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	userOpIndexInArgs = 1
)

type UserOperation struct {
	// TODO: check types, verify nothing lost in conversion
	To        string `json:"to"`        // Atlas address
	Call      string `json:"call"`      // UserCall
	Signature string `json:"signature"` // signature
}

/*
struct UserCall {
    address from;
    address to;
    uint256 deadline;
    uint256 gas;
    uint256 nonce;
    uint256 maxFeePerGas;
    uint256 value;
    address control; // address for preOps / validation funcs
    bytes data;
}
*/

func (h *Harness) buildUserOperation(userCallRaw string, privateKey string) (string, error) {

	pk, err := crypto.ToECDSA([]byte(privateKey))
	if err != nil {
		return "", err
	}

	signature, err := crypto.Sign([]byte(userCallRaw), pk)
	if err != nil {
		return "", err
	}

	userOp, err := json.Marshal(UserOperation{
		To:        h.Atlas.getContractAddress().Hex(),
		Call:      userCallRaw,
		Signature: string(signature),
	})
	if err != nil {
		return "", err
	}

	return string(userOp), nil
}
