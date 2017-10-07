package blockchain

import (
	"os"

	fabricTxn "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/sirupsen/logrus"
)

type ChainCodeInterface interface {
	QueryBoardState() (string, error)
	QueryPlayersState() (string, error)
	PlayerAction(name, num string) (string, error)
}

func NewChainCode() *BaseSetup {
	baseSetup := BaseSetup{
		ConfigFile:      ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   os.Getenv("GOPATH") + "/src/github.com/wadelee1986/guessNumByExample/src/blockchainSDK/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := baseSetup.Initialize(); err != nil {
		logrus.Fatalf(err.Error())
	}

	if err := baseSetup.installAndInstantiateGuessNumCC(); err != nil {
		logrus.Fatalf("installAndInstantiateGuessNumCC return error: %v", err)
	}

	return &baseSetup
}

// installAndInstantiateGuessNumCC ..
func (setup *BaseSetup) installAndInstantiateGuessNumCC() error {

	chainCodePath := "guessNum"
	chainCodeVersion := "v0"

	if setup.ChainCodeID == "" {
		setup.ChainCodeID = GenerateRandomID()
	}

	if err := setup.InstallCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, nil); err != nil {
		return err
	}

	var args []string
	args = append(args, "init")
	args = append(args, "100")

	return setup.InstantiateCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, args)
}

func (setup *BaseSetup) PlayerAction(name, num string) (string, error) {
	fcn := "invoke"

	var args []string
	args = append(args, "playeraction")
	args = append(args, name)
	args = append(args, num)

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	txn, err := fabricTxn.InvokeChaincode(setup.Client, setup.Channel, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, setup.EventHub, setup.ChainCodeID, fcn, args, transientDataMap)
	format := "palyer action return txn id: %v  nonce: %v"
	logrus.Debugf(format, txn.ID, txn.Nonce)
	return "", err
}

func (setup *BaseSetup) QueryPlayersState() (string, error) {
	fcn := "invoke"
	var args []string
	args = append(args, "queryplayersstate")
	return setup.Query(setup.ChannelID, setup.ChainCodeID, fcn, args)
}

// QueryBoardState ...
func (setup *BaseSetup) QueryBoardState() (string, error) {
	fcn := "invoke"
	var args []string
	args = append(args, "queryboardstate")
	return setup.Query(setup.ChannelID, setup.ChainCodeID, fcn, args)
}
