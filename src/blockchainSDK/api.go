package blockchain

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	fabricTxn "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	admin "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/admin"
	"github.com/sirupsen/logrus"
)

type ChainCodeInterface interface {
	QueryValue() (string, error)
	QueryBoardState() (string, error)
	PlayerAction(name, num string) (string, error)
}

var org1Name = "Org1"

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetup struct {
	Client          fab.FabricClient
	Channel         fab.Channel
	EventHub        fab.EventHub
	ConnectEventHub bool
	ConfigFile      string
	OrgID           string
	ChannelID       string
	ChainCodeID     string
	Initialized     bool
	ChannelConfig   string
	AdminUser       ca.User
}

func (setup *BaseSetup) Initialize() error {

	sdkOptions := deffab.Options{
		ConfigFile: setup.ConfigFile,
		//		OrgID:      setup.OrgID,
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/enroll_user",
		},
	}

	sdk, err := deffab.NewSDK(sdkOptions)
	if err != nil {
		return fmt.Errorf("Error initializing SDK: %s", err)
	}

	session, err := sdk.NewPreEnrolledUserSession(setup.OrgID, "Admin")
	if err != nil {
		return fmt.Errorf("Error getting admin user session for org: %s", err)
	}

	sc, err := sdk.NewSystemClient(session)
	if err != nil {
		return fmt.Errorf("NewSystemClient returned error: %v", err)
	}

	setup.Client = sc
	setup.AdminUser = session.Identity()

	channel, err := setup.GetChannel(setup.Client, setup.ChannelID, []string{setup.OrgID})
	if err != nil {
		return fmt.Errorf("Create channel (%s) failed: %v", setup.ChannelID, err)
	}
	setup.Channel = channel

	ordererAdmin, err := sdk.NewPreEnrolledUser("ordererorg", "Admin")
	if err != nil {
		return fmt.Errorf("Error getting orderer admin user: %v", err)
	}

	// Check if primary peer has joined channel
	alreadyJoined, err := HasPrimaryPeerJoinedChannel(sc, channel)
	if err != nil {
		return fmt.Errorf("Error while checking if primary peer has already joined channel: %v", err)
	}

	if !alreadyJoined {
		// Create, initialize and join channel
		if err = admin.CreateOrUpdateChannel(sc, ordererAdmin, setup.AdminUser, channel, setup.ChannelConfig); err != nil {
			return fmt.Errorf("CreateChannel returned error: %v", err)
		}
		time.Sleep(time.Second * 3)

		if err = channel.Initialize(nil); err != nil {
			return fmt.Errorf("Error initializing channel: %v", err)
		}

		if err = admin.JoinChannel(sc, setup.AdminUser, channel); err != nil {
			return fmt.Errorf("JoinChannel returned error: %v", err)
		}
	}

	if err := setup.setupEventHub(sc); err != nil {
		return err
	}

	setup.Initialized = true

	return nil

}

// getEventHub initilizes the event hub
func (setup *BaseSetup) getEventHub(client fab.FabricClient) (fab.EventHub, error) {
	eventHub, err := events.NewEventHub(client)
	if err != nil {
		return nil, fmt.Errorf("Error creating new event hub: %v", err)
	}
	foundEventHub := false
	peerConfig, err := client.Config().PeersConfig(setup.OrgID)
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		if p.URL != "" {
			logrus.Debugf("EventHub connect to peer (%s)", p.URL)
			serverHostOverride := ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			eventHub.SetPeerAddr(p.EventURL, p.TLSCACerts.Path, serverHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found")
	}

	return eventHub, nil
}

func (setup *BaseSetup) setupEventHub(client fab.FabricClient) error {
	eventHub, err := setup.getEventHub(client)
	if err != nil {
		return err
	}

	if setup.ConnectEventHub {
		if err := eventHub.Connect(); err != nil {
			return fmt.Errorf("Failed eventHub.Connect() [%s]", err)
		}
	}
	setup.EventHub = eventHub

	return nil
}

// GetChannel initializes and returns a channel based on config
func (setup *BaseSetup) GetChannel(client fab.FabricClient, channelID string, orgs []string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel return error: %v", err)
	}

	ordererConfig, err := client.Config().RandomOrdererConfig()
	if err != nil {
		return nil, fmt.Errorf("RandomOrdererConfig() return error: %s", err)
	}
	serverHostOverride := ""
	if str, ok := ordererConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}
	orderer, err := orderer.NewOrderer(ordererConfig.URL, ordererConfig.TLSCACerts.Path, serverHostOverride, client.Config())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, fmt.Errorf("Error adding orderer: %v", err)
	}

	for _, org := range orgs {
		peerConfig, err := client.Config().PeersConfig(org)
		if err != nil {
			return nil, fmt.Errorf("Error reading peer config: %v", err)
		}
		for _, p := range peerConfig {
			serverHostOverride = ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			endorser, err := deffab.NewPeer(p.URL, p.TLSCACerts.Path, serverHostOverride, client.Config())
			if err != nil {
				return nil, fmt.Errorf("NewPeer return error: %v", err)
			}
			err = channel.AddPeer(endorser)
			if err != nil {
				return nil, fmt.Errorf("Error adding peer: %v", err)
			}
		}
	}

	return channel, nil
}

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return randomString(10)
}

// InstallAndInstantiateExampleCC ..
func (setup *BaseSetup) InstallAndInstantiateExampleCC() error {

	chainCodePath := "chaincode"
	chainCodeVersion := "v0"

	if setup.ChainCodeID == "" {
		setup.ChainCodeID = GenerateRandomID()
	}

	if err := setup.InstallCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, nil); err != nil {
		return err
	}

	var args []string
	args = append(args, "init")
	//args = append(args, "a")
	args = append(args, "100")
	//args = append(args, "b")
	//args = append(args, "200")

	return setup.InstantiateCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, args)
}

// InstallAndInstantiateExampleCC ..
func (setup *BaseSetup) InstallAndInstantiateExampleCCBak() error {

	chainCodePath := "chaincode"
	chainCodeVersion := "v0"

	if setup.ChainCodeID == "" {
		setup.ChainCodeID = GenerateRandomID()
	}

	if err := setup.InstallCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, nil); err != nil {
		return err
	}

	var args []string
	args = append(args, "init")
	args = append(args, "a")
	args = append(args, "100")
	args = append(args, "b")
	args = append(args, "200")

	return setup.InstantiateCC(setup.ChainCodeID, chainCodePath, chainCodeVersion, args)
}

// QueryAsset ...
func (setup *BaseSetup) QueryAsset() (string, error) {
	fcn := "invoke"
	var args []string
	args = append(args, "query")
	args = append(args, "b")
	return setup.Query(setup.ChannelID, setup.ChainCodeID, fcn, args)
}

// Query ...
func (setup *BaseSetup) Query(channelID string, chainCodeID string, fcn string, args []string) (string, error) {
	return fabricTxn.QueryChaincode(setup.Client, setup.Channel, chainCodeID, fcn, args)
}

// InstantiateCC ...
func (setup *BaseSetup) InstantiateCC(chainCodeID string, chainCodePath string, chainCodeVersion string, args []string) error {

	chaincodePolicy := cauthdsl.SignedByMspMember(setup.Client.UserContext().MspID())

	return admin.SendInstantiateCC(setup.Channel, chainCodeID, args, chainCodePath, chainCodeVersion, chaincodePolicy, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, setup.EventHub)
}

// InstallCC ...
func (setup *BaseSetup) InstallCC(chainCodeID string, chainCodePath string, chainCodeVersion string, chaincodePackage []byte) error {

	if err := admin.SendInstallCC(setup.Client, chainCodeID, chainCodePath, chainCodeVersion, chaincodePackage, setup.Channel.Peers(), setup.GetDeployPath()); err != nil {
		return fmt.Errorf("SendInstallProposal return error: %v", err)
	}

	return nil
}

// GetDeployPath ..
func (setup *BaseSetup) GetDeployPath() string {
	pwd, _ := os.Getwd()
	return path.Join(pwd, "../fixtures/testdata")
}

func NewChainCode() *BaseSetup {
	baseSetup := BaseSetup{
		ConfigFile:      ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "/home/wade.lee/goWorkProject/src/github.com/hyperledger/firstcc/server/src/blockchain/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := baseSetup.Initialize(); err != nil {
		logrus.Fatalf(err.Error())
	}

	if err := baseSetup.InstallAndInstantiateExampleCC(); err != nil {
		logrus.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	return &baseSetup
}

func (setup *BaseSetup) QueryValue() (string, error) {

	// Get Query value before invoke
	value, err := setup.QueryAsset()
	if err != nil {
		logrus.Fatalf("getQueryValue return error: %v", err)
		return value, err
	}
	logrus.Debugf("*** QueryValue before invoke %s", value)
	newMoveFunds(setup)

	value, err = setup.QueryAsset()
	if err != nil {
		logrus.Fatalf("getQueryValue return error: %v", err)
		return value, err
	}
	logrus.Debugf("*** QueryValue after invoke %s", value)

	return value, err
}

// moveFunds ...
func newMoveFunds(setup *BaseSetup) error {
	fcn := "invoke"

	var args []string
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	_, err := fabricTxn.InvokeChaincode(setup.Client, setup.Channel, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, setup.EventHub, setup.ChainCodeID, fcn, args, transientDataMap)
	return err
}

func (setup *BaseSetup) PlayerAction(name, num string) (string, error) {
	fcn := "invoke"

	var args []string
	args = append(args, "playeraction")
	args = append(args, name)
	args = append(args, num)

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	_, err := fabricTxn.InvokeChaincode(setup.Client, setup.Channel, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, setup.EventHub, setup.ChainCodeID, fcn, args, transientDataMap)
	return "", err
}

// QueryBoardState ...
func (setup *BaseSetup) QueryBoardState() (string, error) {
	fcn := "invoke"
	var args []string
	args = append(args, "queryboardstate")
	return setup.Query(setup.ChannelID, setup.ChainCodeID, fcn, args)
}
