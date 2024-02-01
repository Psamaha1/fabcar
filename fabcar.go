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

type DataFF struct {
	Tempff      string `json:"tempff"`
	Timestampff string `json:"timestampff"`
	Causeff     string `json:"causeff"`
}

type Transaction struct {
	Sender      string `json:"sender"`
	Receiver    string `json:"receiver"`
	TimestampT  string `json:"timestampt"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
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
	case "createDefectFF":
		return s.createDefectFF(APIstub, args)
	case "createTransaction":
		return s.createTransaction(APIstub, args)
	case "queryAllDefects":
		return s.queryAllDefects(APIstub)
	case "queryAllSensors":
		return s.queryAllSensors(APIstub)
	case "queryAllTransactions":
		return s.queryAllTransactions(APIstub)
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

func (s *SmartContract) createDefectFF(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("The number of arguments doesnt look right ma friend, Expecting 3")
	}
	var data = DataFF{Tempff: args[1], Timestampff: args[2], Causeff: args[3]}
	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(args[0], dataAsBytes)
	indexName := "causeff~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{data.Causeff, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)
	return shim.Success(dataAsBytes)
}

func (s *SmartContract) createTransaction(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 6 {
		return shim.Error("The number of arguments doesnt look right ma friend, Expecting 5")
	}
	var data = Transaction{Sender: args[1], Receiver: args[2], TimestampT: args[3], Amount: args[4], Description: args[5]}
	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(args[0], dataAsBytes)
	indexName := "sender~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{data.Sender, args[0]})
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

func (s *SmartContract) queryAllDefects(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "Shipment0"
	endKey := "Shipment999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllDefects:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryAllSensors(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "Sensor0"
	endKey := "Sensor999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllSensors:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryAllTransactions(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "Transaction0"
	endKey := "Transaction999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllTransaction:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
