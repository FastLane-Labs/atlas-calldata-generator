package contracts

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/log"
)

var (
	ErrCantFindMethodABI      = errors.New("err - cant find method in ABI")
	ErrCantFindStructFieldABI = errors.New("err - cant find struct field in ABI")
	ErrStructFieldOutOfRange  = errors.New("err - struct field index out of range")
	ErrCantFindMethodJSON     = errors.New("err - cant find method in JSON tx args")
	ErrNotDynamicType         = errors.New("err - not dynamic type")
	ErrCantAssignField        = errors.New("err - cant assign source field to dest field")
	ErrCantSetDestField       = errors.New("err - cant set dest field")
	ErrCantDecodeBytesArray   = errors.New("err - cant decode to intended bytes array format")
	ErrBytesSizeOverTarget    = errors.New("err - bytes array size exceeds target")
)

func (c *Contract) EncodeTxData(funcName string, funcArgs string) ([]byte, error) {
	// funcArgs must be a JSON-format string
	c.Lock.RLock()
	defer c.Lock.RUnlock()

	method, ok := c.ABI.Methods[funcName]
	if !ok {
		return []byte{0}, ErrCantFindMethodABI
	}

	argValues, err := handleLayerZeroUnmarshal(funcArgs)
	if err != nil {
		log.Warn("err", "encodeTx 0a", err)
		return []byte{0}, err
	}

	values, err := buildInterfaceValues(argValues, method)
	if err != nil {
		log.Warn("err", "encodeTx 1a", err)
		return []byte{0}, err
	}

	return c.ABI.Pack(funcName, values...)
}

// indirect recursively dereferences the value until it either gets the value
// or finds a big.Int
// copied from reflect package
func indirect(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr && v.Elem().Type() != reflect.TypeOf(big.Int{}) {
		return indirect(v.Elem())
	}
	return v
}

func handleTuple(data json.RawMessage, t abi.Type) (interface{}, error) {
	newMap := make(map[string]json.RawMessage, len(t.TupleElems))
	if err := json.Unmarshal(data, &newMap); err != nil {
		log.Warn("err - encodeTx 0.b", "could not unmarshal tuple to map", err)
		return handleTupleNoKeys(data, t) // try slice if it isn't a map
	}

	newTuple := indirect(reflect.New(t.GetType()))
	for i, fieldName := range t.TupleRawNames {
		if _, ok := newMap[fieldName]; !ok {
			log.Warn("err - encodeTx 1.b.1", "index", i, "missing field name", fieldName)
			for _fieldName := range newMap {
				log.Warn("err - encodeTx 1.b.2", "existing field name", _fieldName)
			}
			return nil, ErrCantFindStructFieldABI
		}

		newValue, err := handleNestedUnmarshal(newMap[fieldName], *t.TupleElems[i])
		if err != nil {
			return nil, err
		}

		if i >= newTuple.NumField() {
			log.Warn("err - encodeTx 2.b.1", "index", i, "missing field name", fieldName)
			for _fieldName := range newMap {
				log.Warn("err - encodeTx 2.b.2", "existing field name", _fieldName)
			}
			return nil, ErrStructFieldOutOfRange
		}

		dst := newTuple.Field(i)
		if !dst.CanSet() {
			log.Warn("err - encodeTx 3.b.1", "index", i, "field name", fieldName, "err", ErrCantSetDestField)
			return nil, ErrCantSetDestField
		}

		src := indirect(reflect.ValueOf(newValue))
		if !src.Type().AssignableTo(dst.Type()) {
			log.Warn("err - encodeTx 3.b.2", "index", i, "field name", fieldName, "srcType", src.Type().Name(), "dstType", dst.Type().Name())
			return nil, ErrCantAssignField
		}

		dst.Set(src)
	}

	return newTuple.Interface(), nil
}

func handleTupleNoKeys(data json.RawMessage, t abi.Type) (interface{}, error) {
	var newSlice []json.RawMessage
	if err := json.Unmarshal(data, &newSlice); err != nil {
		log.Warn("err - encodeTx 4.b", "name", t.GetType().Name(), "err", err)
		return nil, err
	}

	newTuple := indirect(reflect.New(t.GetType()))
	for i, fieldName := range t.TupleRawNames {
		newValue, err := handleNestedUnmarshal(newSlice[i], *t.TupleElems[i])
		if err != nil {
			return nil, err
		}

		if i >= newTuple.NumField() {
			log.Warn("err - encodeTx 2.b.1", "index", i, "missing field name", fieldName)
			return nil, ErrStructFieldOutOfRange
		}

		dst := newTuple.Field(i)
		if !dst.CanSet() {
			log.Warn("err - encodeTx 3.b.1", "index", i, "field name", fieldName, "err", ErrCantSetDestField)
			return nil, ErrCantSetDestField
		}

		src := indirect(reflect.ValueOf(newValue))
		if !src.Type().AssignableTo(dst.Type()) {
			log.Warn("err - encodeTx 3.b.2", "index", i, "field name", fieldName, "srcType", src.Type().Name(), "dstType", dst.Type().Name())
			return nil, ErrCantAssignField
		}

		dst.Set(src)
	}

	return newTuple.Interface(), nil
}

func handleSlice(data json.RawMessage, t abi.Type) (interface{}, error) {
	var newSlice []json.RawMessage
	if err := json.Unmarshal(data, &newSlice); err != nil {
		log.Warn("err - encodeTx 4.b", "name", t.GetType().Name(), "err", err)
		return nil, err
	}

	newValueSlice := reflect.MakeSlice(t.GetType(), len(newSlice), len(newSlice))

	for i, value := range newSlice {
		newValue, err := handleNestedUnmarshal(value, *t.Elem)
		if err != nil {
			log.Warn("err - encodeTx 5.b", "index", i, "name", t.GetType().Name(), "err", err)
			return nil, err
		}

		dst := newValueSlice.Index(i)
		if !dst.CanSet() {
			log.Warn("err - encodeTx 6.b.1", "index", i, "err", ErrCantSetDestField)
			return nil, ErrCantSetDestField
		}

		src := indirect(reflect.ValueOf(newValue))
		if !src.Type().AssignableTo(dst.Type()) {
			log.Warn("err - encodeTx 6.b.2", "index", i, "srcType", src.Type().Name(), "dstType", dst.Type().Name())
			return nil, ErrCantAssignField
		}

		dst.Set(src)
	}
	return newValueSlice.Interface(), nil
}

func handleBytes(data json.RawMessage, t abi.Type) (interface{}, error) {
	var targetStr string
	if err := json.Unmarshal(data, &targetStr); err == nil {
		if len(targetStr) >= 2 && targetStr[:2] == "0x" {
			if hexbytes, err := hex.DecodeString(targetStr[2:]); err == nil {
				if value, err := assignBytes(hexbytes, t); err == nil {
					return value, nil
				}
			}
		}
		if hexbytes, err := hex.DecodeString(targetStr); err == nil {
			if value, err := assignBytes(hexbytes, t); err == nil {
				return value, nil
			}
		}
		if value, err := assignBytes([]byte(targetStr), t); err == nil {
			return value, nil
		}
	}
	log.Warn("err -", "couldnt decode bytes, attempting default decode of target", targetStr)
	return handleDefault(data, t)
}

func assignBytes(targetByteSlice []byte, t abi.Type) (interface{}, error) {
	// handle byte slices
	if t.T == abi.BytesTy {
		destValue := indirect(reflect.New(reflect.TypeOf([]byte{0})))
		destValue.SetBytes(targetByteSlice)
		return destValue.Interface(), nil
	}

	// if not a slice, handle array
	arrayType := reflect.ArrayOf(t.Size, reflect.TypeOf(uint8(0)))
	destValuePointer := reflect.New(arrayType)
	destValueActual := indirect(destValuePointer)

	for i := 0; i < t.Size; i++ {
		if i < len(targetByteSlice) {
			destValueActual.Index(i).Set(reflect.ValueOf(uint8(targetByteSlice[i])))
		} else {
			destValueActual.Index(i).Set(reflect.ValueOf(uint8(0)))
		}
	}
	return destValueActual.Interface(), nil
}

func handleDefault(data json.RawMessage, t abi.Type) (interface{}, error) {
	value := reflect.New(t.GetType()).Interface()
	if err := json.Unmarshal(data, &value); err != nil {
		log.Warn("err - encodeTx 7.b - default unmarshal fail", "argType", t.GetType().Name())
		return value, err
	}
	return value, nil
}

func handleNestedUnmarshal(data json.RawMessage, t abi.Type) (interface{}, error) {
	switch t.T {
	case abi.TupleTy:
		return handleTuple(data, t) // tuple -> struct
	case abi.SliceTy, abi.ArrayTy:
		return handleSlice(data, t)
	case abi.BytesTy, abi.FixedBytesTy:
		return handleBytes(data, t)
	default:
		return handleDefault(data, t)
	}
}

func handleLayerZeroUnmarshal(funcArgs string) ([]json.RawMessage, error) {
	var argValues []json.RawMessage
	err := json.Unmarshal([]byte(funcArgs), &argValues)
	return argValues, err
}

func buildInterfaceValues(argValues []json.RawMessage, method abi.Method) ([]interface{}, error) {
	values := make([]interface{}, len(argValues))
	for i := 0; i < len(method.Inputs); i++ {
		_value, _type := argValues[i], method.Inputs[i].Type
		value, err := handleNestedUnmarshal(_value, _type)
		if err != nil {
			log.Warn("err - encodeTx 1a", "index", i, "name", method.Inputs[i].Name, "err", err)
			return values, err
		}
		values[i] = value
	}
	return values, nil
}
