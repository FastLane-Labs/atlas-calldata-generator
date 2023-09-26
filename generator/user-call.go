package generator

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type UserCall struct {
	From         common.Address `json:"from"`         // User EOA
	To           common.Address `json:"to"`           // DApp user is interacting with
	Deadline     *big.Int       `json:"deadline"`     // block.number execution deadline
	Gas          *big.Int       `json:"gas"`          // gas limit for the user's call.
	Nonce        *big.Int       `json:"nonce"`        // user's nonce on Atlas.
	MaxFeePerGas *big.Int       `json:"maxFeePerGas"` // maxFeePerGas
	Value        *big.Int       `json:"value"`        // the msg.value of the user's operation
	Control      common.Address `json:"control"`      // the address of the DAppControl contract.
	Data         string         `json:"data"`         // User's DApp-specific calldata
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

func (h *Harness) BuildUserCall(userCall UserCall) (string, error) {

	userOp, err := json.Marshal(userCall)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(userOp), nil
}
