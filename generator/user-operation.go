package generator

import (
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type UserOperation struct {
	// TODO: check types, verify nothing lost in conversion
	To        string `json:"to"`        // Atlas address
	Call      string `json:"call"`      // UserCall
	Signature string `json:"signature"` // signature
}

/*
struct UserOperation {
    address to; // Atlas
    UserCall call;
    bytes signature;
}
*/

func (h *Harness) BuildUserOperation(userOperation UserOperation) (string, error) {

	if h.Atlas.GetContractAddress().Hex() != userOperation.To {
		return "", errors.New("err - incorrect Atlas address")
	}

	userOp, err := json.Marshal(userOperation)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(userOp), nil
}
