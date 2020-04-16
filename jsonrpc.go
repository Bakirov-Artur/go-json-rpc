package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

type Rpc interface {
	Request([]byte) error
	Response(interface{}) ([]byte, error)
}

type JsonRpcMethod struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      interface{} `json:"id"`
}

type JsonRpcResult struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Id      interface{} `json:"id"`
}

type JsonRpcError struct {
	Jsonrpc string            `json:"jsonrpc"`
	Error   *JsonRpcErrorCode `json:"error"`
	Id      interface{}       `json:"id"`
}

type JsonRpcErrorCode struct {
	Code    int16  `json:"code"`
	Message string `json:"message"`
}

const (
	//http://golang-basic.blogspot.com/2014/07/step-by-step-guide-to-declaring-enums.html
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
	ServerError    = -32000
)

func JsonRpcPrase(data []byte) (*JsonRpc, error) {
	var jrpc JsonRpc
	err := json.Unmarshal(data, &jrpc.method)
	if err != nil {
		return nil, err
	}
	return &jrpc, nil
}

func JsonRpcPrasePramas(data []byte, params interface{}) (*JsonRpc, error) {
	var jrpc = JsonRpc{method: JsonRpcMethod{Params: params}}
	err := json.Unmarshal(data, &jrpc)
	if err != nil {
		return nil, err
	}
	return &jrpc, nil
}

func JsonRpcErrorBuild(code int16, message string) ([]byte, error) {
	var jrpc JsonRpc
	return json.Marshal(jrpc.NewError(code, message))
}

type JsonRpc struct {
	method   JsonRpcMethod
	rpcError JsonRpcError
}

func (j *JsonRpc) SetParams(params interface{}) {
	j.method.Params = params
}
func (j *JsonRpc) NewResult(result interface{}) *JsonRpcResult {
	return &JsonRpcResult{Jsonrpc: "2.0", Result: result, Id: j.method.Id}
}
func (j *JsonRpc) Request(body []byte) error {
	return json.Unmarshal(body, &j.method)
}

func (j *JsonRpc) Response(t interface{}) ([]byte, error) {
	return json.Marshal(t)
}

func (j *JsonRpc) IsNotify() bool {
	if j.method.Id != nil {
		return true
	}
	return false
}

func (j *JsonRpc) NewError(code int16, message string) *JsonRpcError {
	rpcError := JsonRpcErrorCode{code, message}
	return &JsonRpcError{Jsonrpc: "2.0", Error: &rpcError, Id: j.method.Id}
}

func (j *JsonRpc) NewNilError() *JsonRpcError {
	return &JsonRpcError{Jsonrpc: "2.0", Id: j.method.Id}
}

func (j *JsonRpc) Method() JsonRpcMethod {
	return j.method
}
func (j *JsonRpc) ValueToInt(v interface{}) (int64, error) {
	// //type Index struct { index int64 }
	// //indx := Index(params)
	// //x.(type)
	// //Парсим значение params, оставляем только нужный тип
	log.Println("Value type of: ", reflect.TypeOf(v))
	log.Println("Value: ", reflect.ValueOf(v))
	switch reflect.ValueOf(v).Kind() {
	case reflect.Float64, reflect.Float32:
		index := reflect.ValueOf(v)
		log.Println("Value type of Float to int: ", index)
		return int64(index.Interface().(float64)), nil
	case reflect.Int32, reflect.Int64:
		index := reflect.ValueOf(v)
		log.Println("Value type of Int to int64: ", index)
		return int64(index.Interface().(int64)), nil
	case reflect.String:
		index := reflect.ValueOf(v)
		i := index.String()
		log.Println("Value type of String to int64: ", i)
		n, err := strconv.Atoi(i)
		if err != nil {
			return -1, err
		}
		return int64(n), nil
	default:
		return -1, errors.New("Invalid convert Value to Int64")
	}
}

func (j *JsonRpc) ValueToString(v reflect.Value) (string, error) {
	switch v.Elem().Kind() {
	case reflect.String:
		log.Println("Value type of String to String: ", v.Elem())
		return v.Elem().String(), nil
	case reflect.Int32, reflect.Int64:
		log.Println("Value type of Int to String: ", v.Elem())
		return strconv.FormatInt(v.Elem().Int(), 64), nil
	case reflect.Float64, reflect.Float32:
		log.Println("Value type of Float to String: ", v.Elem())
		return fmt.Sprintf("%f", v.Elem().Float()), nil
	default:
		return "", errors.New(fmt.Sprintf("Invalid convert Value (%v) to String ", v.Elem().Kind()))
	}
}

func (j *JsonRpc) ValueToFloat(v reflect.Value) (float64, error) {
	switch v.Elem().Kind() {
	case reflect.String:
		log.Println("Value type of String to Float: ", v.Elem())
		return strconv.ParseFloat(v.Elem().String(), 64)
	case reflect.Int32, reflect.Int64:
		log.Println("Value type of Int to Float: ", v.Elem())
		return float64(v.Elem().Int()), nil
	case reflect.Float64, reflect.Float32:
		log.Println("Value type of Float to Float: ", v.Elem())
		return v.Elem().Float(), nil
	default:
		return -0.1, errors.New(fmt.Sprintf("Invalid convert Value (%v) to Float ", v.Elem().Kind()))
	}
}

func (j *JsonRpc) SliceUnPack(vars ...interface{}) error {

	vSliceItems := reflect.ValueOf(j.method.Params)
	vVarItems := reflect.ValueOf(vars)

	if vSliceItems.Kind() != reflect.Slice && vVarItems.Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("Invalid type wait Slice, have type: %v", vVarItems.Kind()))
	}

	if vSliceItems.Len() != vVarItems.Len() {
		return errors.New(fmt.Sprintf("Invalid size Slice: ", vSliceItems.Len(), " < params - vars > ", vVarItems.Len()))
	}

	for i := 0; i < vSliceItems.Len(); i++ {
		sItem := vSliceItems.Index(i)
		if sItem.Elem().Kind() == reflect.Ptr {
			sItem = sItem.Elem().Elem()
			log.Println("sItem kind(): ", sItem.Kind())
			log.Println("sItem value: ", sItem)
		}

		vItem := vVarItems.Index(i)
		if vItem.Elem().Kind() == reflect.Ptr {
			vItem = vItem.Elem().Elem()
			log.Println("vItem kind(): ", vItem.Kind())
			log.Println("vItem value: ", vItem)
		}

		// log.Println("vSliceItem item Kind:", reflect.ValueOf(sItem).Kind())
		// log.Println("vSliceItem item Elem():", sItem.Elem())
		// log.Println("vSliceItem item Elem() kind:", sItem.Elem().Kind())

		//Конвертирование sItem в тип vItem
		if !vItem.CanSet() {
			return errors.New(fmt.Sprintf("Invalid change variables: ", vItem, ", kind", vItem.Elem().Kind()))
		}
		switch vItem.Kind() {
		case reflect.String:
			s, e := j.ValueToString(sItem)
			if e != nil {
				return e
			}
			vItem.SetString(s)
		case reflect.Int32, reflect.Int64:
			i, e := j.ValueToInt(sItem.Interface())
			if e != nil {
				return e
			}
			vItem.SetInt(i)
		case reflect.Float64, reflect.Float32:
			f, e := j.ValueToFloat(sItem)
			if e != nil {
				return e
			}
			vItem.SetFloat(f)
		default:
			return errors.New(fmt.Sprintf("Unknown type change variables: ", vItem, ", kind", vItem.Kind()))
		}
	}
	// return errors.New("Invalid set var:")
	return nil
}

func (j *JsonRpc) GetValues(v ...interface{}) error {
	params := j.method.Params

	switch reflect.ValueOf(params).Kind() {
	case reflect.Slice:
		return j.SliceUnPack(v...)
	case reflect.Map:
		return j.MapUnPack(v...)
	default:
		return errors.New(fmt.Sprintf("Unknown Params: ", params, ", kind", reflect.ValueOf(params).Kind()))
	}
	return nil
}

func (j *JsonRpc) MapUnPack(vars ...interface{}) error {
	//https://coderoad.ru/26287173/Получение-поля-типа-reflect-Ptr-в-структуре-Go
	//https://stackoverflow.com/questions/24348184/get-pointer-to-value-using-reflection
	//https://play.golang.org/p/nleA2YWMj8
	//log.Println("mapUnPack ", vMap, vars)

	vVarItems := reflect.ValueOf(vars)
	vMapItems := reflect.ValueOf(j.method.Params)

	if vMapItems.Kind() != reflect.Map && vVarItems.Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("Invalid type wait Slice, have type: %v", vVarItems.Kind()))
	}

	if vVarItems.Len() != vMapItems.Len() {
		return errors.New(fmt.Sprintf("Invalid size Map: ", vMapItems.Len(), " < params - vars > ", vVarItems.Len()))
	}

	vMapKeys := vMapItems.MapKeys()
	for vIndex, vKey := range vMapKeys {
		vItems := vVarItems.Index(vIndex)
		vMapIndex := vMapItems.MapIndex(vKey)
		// log.Println("Map key index: ", vIndex)
		// log.Println("vVarItems.Index: ", vItems)
		// log.Println(vKey.String(), vMapIndex.Interface())
		evi := vItems.Elem()
		if evi.Kind() == reflect.Ptr {
			evi = evi.Elem()
		}
		if !evi.CanSet() {
			return errors.New(fmt.Sprintf("Invalid change variables: ", evi, ", kind", evi.Kind()))
		}
		//Тут нужно проверять все небхдимые соответсвия типов
		log.Println("vMapIndex.Elem() kind(): ", vMapIndex.Elem().Kind())
		log.Println("evi kind: ", evi.Kind())
		// vMapIndex = vMapIndex.Elem()
		switch evi.Kind() {
		case reflect.String:
			s, e := j.ValueToString(vMapIndex)
			if e != nil {
				return e
			}
			evi.SetString(s)
		case reflect.Int64, reflect.Int32:
			i, e := j.ValueToInt(vMapIndex.Interface())
			if e != nil {
				return e
			}
			evi.SetInt(i)
		case reflect.Float64, reflect.Float32:
			f, e := j.ValueToFloat(vMapIndex)
			if e != nil {
				return e
			}
			evi.SetFloat(f)
		default:
			return errors.New(fmt.Sprintf("Unknown type change variables: ", evi, ", kind", evi.Kind()))
		}
	}
	return nil
}

func unpack(s []string, vars ...*string) {
	for i, str := range s {
		*vars[i] = str
	}
}

//Syntax:
//
//--> data sent to Server

//<-- data sent to Client
//rpc call with positional parameters:
//
//--> {"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}
//<-- {"jsonrpc": "2.0", "result": 19, "id": 1}
//
//--> {"jsonrpc": "2.0", "method": "subtract", "params": [23, 42], "id": 2}
//<-- {"jsonrpc": "2.0", "result": -19, "id": 2}
//rpc call with named parameters:
//
//--> {"jsonrpc": "2.0", "method": "subtract", "params": {"subtrahend": 23, "minuend": 42}, "id": 3}
//<-- {"jsonrpc": "2.0", "result": 19, "id": 3}
//
//--> {"jsonrpc": "2.0", "method": "subtract", "params": {"minuend": 42, "subtrahend": 23}, "id": 4}
//<-- {"jsonrpc": "2.0", "result": 19, "id": 4}
//a Notification:
//
//--> {"jsonrpc": "2.0", "method": "update", "params": [1,2,3,4,5]}
//--> {"jsonrpc": "2.0", "method": "foobar"}
//rpc call of non-existent method:
//
//--> {"jsonrpc": "2.0", "method": "foobar", "id": "1"}
//<-- {"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "1"}
//rpc call with invalid JSON:
//
//--> {"jsonrpc": "2.0", "method": "foobar, "params": "bar", "baz]
//<-- {"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}
//rpc call with invalid Request object:
//
//--> {"jsonrpc": "2.0", "method": 1, "params": "bar"}
//<-- {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//rpc call Batch, invalid JSON:
//
//--> [
//  {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
//  {"jsonrpc": "2.0", "method"
//]
//<-- {"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}
//rpc call with an empty Array:
//--> []
//<-- {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//rpc call with an invalid Batch (but not empty):
//--> [1]
//<-- [
//  {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//]
//rpc call with invalid Batch:
//
//--> [1,2,3]
//<-- [
//  {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//  {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//  {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
//]
//rpc call Batch:
//
//--> [
//        {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
//        {"jsonrpc": "2.0", "method": "notify_hello", "params": [7]},
//        {"jsonrpc": "2.0", "method": "subtract", "params": [42,23], "id": "2"},
//        {"foo": "boo"},
//        {"jsonrpc": "2.0", "method": "foo.get", "params": {"name": "myself"}, "id": "5"},
//        {"jsonrpc": "2.0", "method": "get_data", "id": "9"}
//    ]
//<-- [
//        {"jsonrpc": "2.0", "result": 7, "id": "1"},
//        {"jsonrpc": "2.0", "result": 19, "id": "2"},
//        {"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null},
//        {"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "5"},
//        {"jsonrpc": "2.0", "result": ["hello", 5], "id": "9"}
//    ]
//rpc call Batch (all notifications):
//
//--> [
//        {"jsonrpc": "2.0", "method": "notify_sum", "params": [1,2,4]},
//        {"jsonrpc": "2.0", "method": "notify_hello", "params": [7]}
//    ]
//<-- //Nothing is returned for all notification batches
