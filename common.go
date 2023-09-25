package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// build UserData onto this
type Harness struct {
	DAppMap    map[common.Address]*Contract
	AddressMap map[string]common.Address
	Atlas      *Contract
	Lock       sync.Mutex
}

func NewHarness() *Harness {
	addressMap := loadAddresses()
	return &Harness{
		DAppMap:    make(map[common.Address]*Contract),
		AddressMap: loadAddresses(),
		Atlas:      loadAtlas(addressMap["Atlas"]),
	}
}

func (h *Harness) NewContract(contractAddress common.Address, abiRawJSON string) error {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	dappContract, err := NewContract(contractAddress, abiRawJSON)
	if err != nil {
		return err
	}
	methodNames := dappContract.getMethodNames()
	if len(methodNames) < 1 {
		return errors.New("err - no available methods found")
	}
	h.DAppMap[dappContract.getContractAddress()] = dappContract
	return nil
}

func (h *Harness) loadContractOS(dappAddress common.Address, path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var nestedABI map[string]json.RawMessage
	err = json.Unmarshal([]byte(byteValue), &nestedABI)
	if err != nil {
		return err
	}

	err = h.NewContract(dappAddress, string(nestedABI["abi"]))
	if err != nil {
		return err
	}
	return nil
}

func (h *Harness) printArgs(dapp string, funcName string) error {
	dappAddr, ok := h.AddressMap[dapp]
	if !ok {
		return errors.New("err - DApp Address not found")
	}

	contract, ok := h.DAppMap[dappAddr]
	if !ok {
		return errors.New("err - DApp Contract not found")
	}

	fmt.Println("arguments for the " + funcName + " function of " + dapp)
	contract.printArgs(funcName)
	return nil
}

func (h *Harness) generateDAppCalldata(dapp string, funcName string, args string) (string, error) {
	dappAddr, ok := h.AddressMap[dapp]
	if !ok {
		return "", errors.New("err - DApp Address not found A")
	}

	contract, ok := h.DAppMap[dappAddr]
	if !ok {
		return "", errors.New("err - DApp Contract not found B")
	}

	result, err := contract.encodeTxData(funcName, args)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(result), nil
}

func loadAddresses() map[string]common.Address {
	jsonFile, err := os.Open("./config.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var addressMapRaw map[string]string // json.RawMessage
	err = json.Unmarshal([]byte(byteValue), &addressMapRaw)
	if err != nil {
		fmt.Println(err)
	}

	addressMap := make(map[string]common.Address, len(addressMapRaw))
	for name, addr := range addressMapRaw {
		addressMap[name] = common.HexToAddress(addr)
	}
	return addressMap
}

func loadAtlas(atlasAddress common.Address) *Contract {
	jsonFile, err := os.Open("./abis/Atlas.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var nestedABI map[string]json.RawMessage
	err = json.Unmarshal([]byte(byteValue), &nestedABI)
	if err != nil {
		fmt.Println(err)
	}

	atlasContract, err := NewContract(atlasAddress, string(nestedABI["abi"]))
	if err != nil {
		fmt.Println(err)
	}
	return atlasContract
}
