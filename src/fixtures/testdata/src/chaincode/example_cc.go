package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### example_cc Init ###########")
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	max, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Expecting integer value for max holding")
	}

	return InitializeGame(stub, max)
}

// Init ...
func (t *SimpleChaincode) InitBak(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### example_cc Init ###########")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			fmt.Printf("Transient data in 'init' : %s\n", transientData)
			return shim.Success(transientData)
		}
	}
	return shim.Success(nil)

}

// Query ...
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call")
}

// Invoke ...
// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### example_cc Invoke ###########")
	function, args := stub.GetFunctionAndParameters()

	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	if args[0] == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	}

	if args[0] == "query" {
		// queries an entity state
		return t.query(stub, args)
	}
	if args[0] == "queryboardstate" {
		return t.queryBoardState(stub, args)
	}

	if args[0] == "playeraction" {
		return t.playerAction(stub, args)
	}

	if args[0] == "move" {
		if err := stub.SetEvent("testEvent", []byte("Test Payload")); err != nil {
			return shim.Error("Unable to set CC event: testEvent. Aborting transaction ...")
		}
		return t.move(stub, args)
	}
	return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'")
}

func (t *SimpleChaincode) move(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// must be an invoke
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4, function followed by 2 names and 1 value")
	}

	A = args[1]
	B = args[2]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			fmt.Printf("Transient data in 'move' : %s\n", transientData)
			return shim.Success(transientData)
		}
	}
	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[1]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) queryBoardState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("query board state Incorrect number of arguments. Expecting name of the person to query")
	}

	boardAsBytes, err := GetBoardStateBytes(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(boardAsBytes)
}

func (t *SimpleChaincode) playerAction(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		return shim.Error("PlayerAction Incorrect number of arguments. Expecting at least 3")
	}

	board, err := GetBoardState(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	name := args[1]
	action := args[2]

	val, err := strconv.Atoi(action)
	if err != nil {
		return shim.Error("Expecting integer value for action holding")
	}

	if p, ok := board.Players[name]; ok {
		board.Players[name] = Player{
			Balance: p.Balance,
			Bets:    append(p.Bets, val),
		}
	} else {
		pl := Player{
			Balance: 0,
			Bets:    make([]int, 0),
		}
		pl.Bets = append(pl.Bets, val)
		board.Players[name] = pl
	}

	totalBets := 0
	for _, v := range board.Players {
		for _, bet := range v.Bets {
			if bet > 0 {
				totalBets = totalBets + bet
			}
		}
	}

	if totalBets >= board.Max {
		hasWinner, newRoundBoard := FindWinner(board)
		if hasWinner {
			return PutBoardStateByResponse(stub, newRoundBoard, shim.Success([]byte("haswinner")))
		}
		return PutBoardStateByResponse(stub, newRoundBoard, shim.Success([]byte("nowinner")))
	}

	res := PutBoardState(stub, board)
	if transientMap, err := stub.GetTransient(); err == nil {
		if transientData, ok := transientMap["result"]; ok {
			fmt.Printf("Transient data in 'move' : %s\n", transientData)
			return shim.Success(transientData)
		}
	}
	return res
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[1]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
