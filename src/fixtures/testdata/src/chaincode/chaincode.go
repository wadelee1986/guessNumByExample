package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### chain code Init ###########")
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

// Invoke ...
// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### chain code Invoke ###########")
	function, args := stub.GetFunctionAndParameters()

	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	if args[0] == "queryboardstate" {
		return t.queryBoardState(stub, args)
	}

	if args[0] == "queryplayersstate" {
		return t.queryplayersstate(stub, args)
	}

	if args[0] == "playeraction" {
		return t.playerAction(stub, args)
	}

	return shim.Error("Unknown action, check the first argument, must be one of 'queryboardstate', 'queryplayersstate', or 'playeraction'")
}

func (t *SimpleChaincode) queryplayersstate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("queryplayersstate Incorrect number of arguments. Expecting 1")
	}

	board, err := GetBoardState(stub)

	if err != nil {
		return shim.Error(err.Error())
	}

	res := []PlasersState{}

	for name, v := range board.Players {
		res = append(res, PlasersState{
			Name:      name,
			Balance:   v.Balance,
			NumOfBets: len(v.Bets),
		})
	}

	resAsBytes, err := json.Marshal(res)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(resAsBytes)
}

func (t *SimpleChaincode) queryBoardState(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("query board state Incorrect number of arguments. Expecting 1")
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

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
