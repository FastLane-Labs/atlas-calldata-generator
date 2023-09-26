package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/polygon-fast-lane/calldata-generator/contracts"
)

const (
	ABI_DIRECTORY string = "abis"
)

// build UserData onto this
type Harness struct {
	DAppMap    map[common.Address]*contracts.Contract
	AddressMap map[string]common.Address
	Atlas      *contracts.Contract
	Lock       sync.Mutex
}

func NewHarness() *Harness {
	addressMap := loadAddresses()
	return &Harness{
		DAppMap:    make(map[common.Address]*contracts.Contract),
		AddressMap: loadAddresses(),
		Atlas:      loadAtlas(addressMap["Atlas"]),
	}
}

func (h *Harness) Run() {
	for dapp := range h.AddressMap {
		if dapp != "Atlas" {
			err := h.newDAppFromOS(dapp)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(dapp + " DApp Loaded")
			}
		}
	}
}

func (h *Harness) newDAppFromOS(dapp string) error {

	swapIntentAddr, ok := h.AddressMap[dapp]
	if !ok {
		return errors.New("err - " + dapp + " not in config")
	}

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	path := filepath.Join(workDir, ABI_DIRECTORY, dapp+".json")
	fmt.Println(path)
	err = h.loadContractOS(swapIntentAddr, path)
	if err != nil {
		return err
	}
	return nil
}

func (h *Harness) NewContract(contractAddress common.Address, abiRawJSON string) error {
	h.Lock.Lock()
	defer h.Lock.Unlock()

	dappContract, err := contracts.NewContract(contractAddress, abiRawJSON)
	if err != nil {
		return err
	}
	methodNames := dappContract.GetMethodNames()
	if len(methodNames) < 1 {
		return errors.New("err - no available methods found")
	}
	h.DAppMap[dappContract.GetContractAddress()] = dappContract
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
	contract.PrintArgs(funcName)
	return nil
}

func loadAddresses() map[string]common.Address {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	jsonFile, err := os.Open(filepath.Join(workDir, "config.json"))
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

func loadAtlas(atlasAddress common.Address) *contracts.Contract {
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	jsonFile, err := os.Open(filepath.Join(workDir, ABI_DIRECTORY, "Atlas.json"))
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	atlasContract, err := contracts.NewContract(atlasAddress, string(byteValue))
	if err != nil {
		fmt.Println(err)
	}
	return atlasContract
}
