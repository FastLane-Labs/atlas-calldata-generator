package main

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	atlasFuncName string = "metacall"
	atlasName     string = "Atlas"
)

/*
function metacall(
	DAppConfig calldata dConfig, // supplied by frontend
	UserOperation calldata userOp, // set by user
	SolverOperation[] calldata solverOps, // supplied by FastLane via frontend integration
	Verification calldata verification // supplied by front end after it sees the other data
) external payable
*/

type MetacallArgs struct {
	DConfig      string   `json:"dConfig"`   // DAppConfig
	UserOp       string   `json:"userOp"`    // UserOperation
	SolverOps    []string `json:"solverOps"` // SolverOperations
	Verification string   `json:"verification"`
}

func (h *Harness) buildAtlasCalldata(
	dConfig string, userOp string, solverOps []string, verification string) (*hexutil.Bytes, error) {

	calldataRaw, err := json.Marshal(MetacallArgs{
		DConfig:      dConfig,
		UserOp:       userOp,
		SolverOps:    solverOps,
		Verification: verification,
	})
	if err != nil {
		return new(hexutil.Bytes), err
	}

	calldataBytes, err := h.Atlas.encodeTxData(atlasFuncName, string(calldataRaw))
	return (*hexutil.Bytes)(&calldataBytes), nil
}
