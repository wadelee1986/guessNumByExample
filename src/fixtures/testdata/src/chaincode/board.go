package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type Board struct {
	Players map[string]Player `json:"players"`
	Max     int               `json:"max"`
}

func (b *Board) getPlayerByName(name string) Player {
	if v, ok := b.Players[name]; ok {
		return v
	}
	return Player{}
}

func (b *Board) putPlayerByName(name string, pl Player) {
	b.Players[name] = pl
}

func InitializeGame(stub shim.ChaincodeStubInterface, max int) pb.Response {
	// Initialize the chaincode
	board := Board{Max: max,
		Players: make(map[string]Player)}
	return PutBoardState(stub, board)
}

func AddPlayer(stub shim.ChaincodeStubInterface, name string) pb.Response {
	board, err := GetBoardState(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if _, ok := board.Players[name]; ok {
		return shim.Error(fmt.Sprintf("player name: %v  is already existed", name))
	} else {
		board.Players[name] = Player{Bets: make([]int, 0)}
	}
	return PutBoardState(stub, board)
}

func RemovePlayer(stub shim.ChaincodeStubInterface, name string) pb.Response {
	board, err := GetBoardState(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if _, ok := board.Players[name]; ok {
		delete(board.Players, name)
	} else {
		return shim.Error(fmt.Sprintf("player name: %v  is not existed", name))
	}
	return PutBoardState(stub, board)
}

func GetBoardStateBytes(stub shim.ChaincodeStubInterface) ([]byte, error) {
	//Get the board state from the ledger
	boardAsBytes, err := stub.GetState("board")
	if err != nil {
		return nil, errors.New("Could not find board")
	}
	if boardAsBytes == nil {
		return nil, errors.New("Entity not found")
	}
	return boardAsBytes, nil
}

func GetBoardState(stub shim.ChaincodeStubInterface) (Board, error) {
	//Get the board state from the ledger
	var board Board
	boardAsBytes, err := stub.GetState("board")
	if err != nil {
		return Board{}, errors.New("Could not find board")
	}
	if boardAsBytes == nil {
		return Board{}, errors.New("Entity not found")
	}
	err = json.Unmarshal(boardAsBytes, &board)
	if err != nil {
		return Board{}, errors.New("Error Unmarshalling json")
	}
	return board, nil
}

func PutBoardState(stub shim.ChaincodeStubInterface, board Board) pb.Response {

	boardAsBytes, err := json.Marshal(board)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Write the state to the ledger
	err = stub.PutState("board", boardAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func PutBoardStateByResponse(stub shim.ChaincodeStubInterface, board Board, res pb.Response) pb.Response {

	boardAsBytes, err := json.Marshal(board)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Write the state to the ledger
	err = stub.PutState("board", boardAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return res
}

func GetPlayersBlanceAndBetNum(stub shim.ChaincodeStubInterface) {
}
