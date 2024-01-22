package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

type SmartContract struct {
}

type Data struct {
	Temp      string `json:"temp"`
	Hum       string `json:"hum"`
	Tilt      string `json:"tilt"`
	Location  string `json:"location"`
	Timestamp string `json:"timestamp"`
	Cause     string `json:"cause"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabcar_cc")

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))
	switch function {
	case "queryDefect":
		return s.queryDefect(APIstub, args)
	case "createDefect":
		return s.createDefect(APIstub, args)
	case "getHistoryForDefect":
		return s.getHistoryForDefect(APIstub, args)
	case "restictedMethod":
		return s.restictedMethod(APIstub, args)
	default:
		return shim.Error("You sure about the smart contract name? nah, exactly")
	}
}

func (s *SmartContract) queryDefect(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("The number of arguments doesnt look right ma friend, Expecting 1")
	}
	dataAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(dataAsBytes)
}

func (s *SmartContract) createDefect(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 7 {
		return shim.Error("The number of arguments doesnt look right ma friend, Expecting 6")
	}
	var data = Data{Temp: args[1], Hum: args[2], Tilt: args[3], Location: args[4], Timestamp: args[5], Cause: args[6]}
	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(args[0], dataAsBytes)
	indexName := "cause~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{data.Cause, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)
	return shim.Success(dataAsBytes)
}

func (s *SmartContract) restictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		shim.Error("Client identity doesnot posses the attribute")
	}
	if val != "approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as APPROVER have access this method!")
	} else {
		if len(args) != 1 {
			return shim.Error("The number of arguments doesnt look right ma friend, Expecting 1")
		}
		carAsBytes, _ := APIstub.GetState(args[0])
		return shim.Success(carAsBytes)
	}
}

func (t *SmartContract) getHistoryForDefect(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 1 {
		return shim.Error("The number of arguments doesnt look right ma friend Expecting 1")
	}
	dataName := args[0]
	resultsIterator, err := stub.GetHistoryForKey(dataName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
