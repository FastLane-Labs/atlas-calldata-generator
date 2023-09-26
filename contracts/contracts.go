package contracts

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type Contract struct {
	ABI     *abi.ABI
	Address common.Address
	Lock    sync.RWMutex
}

func NewContract(contractAddress common.Address, abiRawJSON string) (*Contract, error) {
	abi, err := getABI(abiRawJSON)
	if err != nil {
		return nil, err
	}
	return &Contract{
		ABI:     &abi,
		Address: contractAddress,
	}, nil
}

func getABI(rawABI string) (abi.ABI, error) {
	var abiString string

	var nestedABI map[string]json.RawMessage
	err := json.Unmarshal([]byte(rawABI), &nestedABI)
	if err != nil {

		var unnestedABI []json.RawMessage
		err2 := json.Unmarshal([]byte(rawABI), &unnestedABI)
		if err2 != nil {
			return abi.ABI{}, errors.New("mapErr " + err.Error() + " arrayErr " + err2.Error())
		}
		abiString = rawABI

	} else {
		if _, ok := nestedABI["abi"]; ok {
			abiString = string(nestedABI["abi"])
		} else {
			abiString = rawABI
		}
	}

	parsedABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		log.Warn("err - loadABI", "error:", err)
		return abi.ABI{}, err
	}
	return parsedABI, nil
}

func (c *Contract) GetMethodNames() []string {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	methodNames := make([]string, len(c.ABI.Methods))
	n := 0
	for name, method := range c.ABI.Methods {
		if method.IsPayable() {
			methodNames[n] = name
			n++
		}
	}
	return methodNames
}

func (c *Contract) PrintArgs(funcName string) {
	if method, ok := c.ABI.Methods[funcName]; ok {
		for i, arg := range method.Inputs {
			fmt.Println(strconv.Itoa(i) + ": " + arg.Name + "  " + arg.Type.String())
		}
	}
}

func (c *Contract) UpdateContractAddress(contractAddress common.Address) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	c.Address = contractAddress
}

func (c *Contract) GetContractAddress() common.Address {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	contractAddress := c.Address
	return contractAddress
}

func (c *Contract) UpdateContractABI(abiRaw []byte) error {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	abi, err2 := abi.JSON(strings.NewReader(string(abiRaw)))
	if err2 != nil {
		log.Warn("err - loadABI 2", "error:", err2)
		return err2
	}
	c.ABI = &abi
	return nil
}
