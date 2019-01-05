package helper

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	logger "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/privval"

	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/contracts/stakemanager"
	"math/big"
)

const (
	NodeFlag               = "node"
	WithHeimdallConfigFlag = "with-heimdall-config"
	HomeFlag               = "home"
	FlagClientHome         = "home-client"
	MainRPCUrl             = "https://kovan.infura.io"
	MaticRPCUrl            = "https://testnet.matic.network"
	CheckpointBufferTime   = time.Second * 256 // aka 256 seconds
)

var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.heimdallcli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.heimdalld")
	MinBalance      = big.NewInt(1000000000000000000) // aka 1 Ether
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, secp256k1.PubKeyAminoRoute, nil)
	cdc.RegisterConcrete(secp256k1.PrivKeySecp256k1{}, secp256k1.PrivKeyAminoRoute, nil)
	Logger = logger.NewTMLogger(logger.NewSyncWriter(os.Stdout))

	//contractCallerObj, err := NewContractCallerObj()
	//if err != nil {
	//	fmt.Errorf("we got error","Error",err)
	//	log.Fatal(err.Error())
	//}
	//
	////app.caller = contractCallerObj
	////caller:= app.caller
	//Logger.Error("contrct","caller",contractCallerObj)
	//value,err:= contractCallerObj.CurrentChildBlock()
	//if err!=nil{
	//	log.Fatal(err.Error())
	//}
	//fmt.Printf("current block",value)
}

// Configuration represents heimdall config
type Configuration struct {
	MainRPCUrl          string `json:"mainRPCUrl"`
	MaticRPCUrl         string `json:"maticRPCUrl"`
	StakeManagerAddress string `json:"stakeManagerAddress"`
	RootchainAddress    string `json:"rootchainAddress"`
	ChildBlockInterval  uint64 `json:"childBlockInterval"`
}

var conf Configuration

// MainChainClient stores eth client for Main chain Network
var mainChainClient *ethclient.Client

// MaticClient stores eth/rpc client for Matic Network
var maticClient *ethclient.Client
var maticRPCClient *rpc.Client

// private key object
var privObject secp256k1.PrivKeySecp256k1
var pubObject secp256k1.PubKeySecp256k1

// Logger stores global logger object
var Logger logger.Logger

// InitHeimdallConfig initializes with viper config (from heimdall configuration)
func InitHeimdallConfig(homeDir string) {
	if strings.Compare(homeDir, "") == 0 {
		// get home dir from viper
		homeDir = viper.GetString(HomeFlag)
	}

	// get heimdall config filepath from viper/cobra flag
	heimdallConfigFilePath := viper.GetString(WithHeimdallConfigFlag)

	// init heimdall with changed config files
	InitHeimdallConfigWith(homeDir, heimdallConfigFilePath)
}

// InitHeimdallConfigWith initializes passed heimdall/tendermint config files
func InitHeimdallConfigWith(homeDir string, heimdallConfigFilePath string) {
	if strings.Compare(homeDir, "") == 0 {
		return
	}

	if strings.Compare(conf.MaticRPCUrl, "") != 0 {
		return
	}

	configDir := filepath.Join(homeDir, "config")
	Logger.Info("Initializing tendermint configurations", "configDir", configDir)

	heimdallViper := viper.New()
	if heimdallConfigFilePath == "" {
		heimdallViper.SetConfigName("heimdall-config") // name of config file (without extension)
		heimdallViper.AddConfigPath(configDir)         // call multiple times to add many search paths
		Logger.Info("Loading heimdall configurations", "file", filepath.Join(configDir, "heimdall-config.json"))
	} else {
		heimdallViper.SetConfigFile(heimdallConfigFilePath) // set config file explicitly
		Logger.Info("Loading heimdall configurations", "file", heimdallConfigFilePath)
	}

	err := heimdallViper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Fatal(err)
	}

	if err = heimdallViper.Unmarshal(&conf); err != nil {
		log.Fatal(err)
	}

	// setup eth client
	if mainChainClient, err = ethclient.Dial(conf.MainRPCUrl); err != nil {
		Logger.Error("Error while creating main chain client", "error", err)
		log.Fatal(err)
	}

	if maticRPCClient, err = rpc.Dial(conf.MaticRPCUrl); err != nil {
		Logger.Error("Error while creating matic chain RPC client", "error", err)
		log.Fatal(err)
	}
	maticClient = ethclient.NewClient(maticRPCClient)

	// load pv file, unmarshall and set to privObject
	privVal := privval.LoadFilePV(filepath.Join(configDir, "priv_validator.json"))
	cdc.MustUnmarshalBinaryBare(privVal.PrivKey.Bytes(), &privObject)
	cdc.MustUnmarshalBinaryBare(privObject.PubKey().Bytes(), &pubObject)

}

// GetConfig returns cached configuration object
func GetConfig() Configuration {
	return conf
}

//
// Root chain
//

func GetRootChainAddress() common.Address {
	return common.HexToAddress(GetConfig().RootchainAddress)
}

func GetRootChainInstance() (*rootchain.Rootchain, error) {
	rootChainInstance, err := rootchain.NewRootchain(GetRootChainAddress(), mainChainClient)
	if err != nil {
		Logger.Error("Unable to create root chain instance", "error", err)
	}

	return rootChainInstance, err
}

func GetRootChainABI() (abi.ABI, error) {
	return abi.JSON(strings.NewReader(rootchain.RootchainABI))
}

//
// Stake manager
//

func GetStakeManagerAddress() common.Address {
	return common.HexToAddress(GetConfig().StakeManagerAddress)
}

func GetStakeManagerInstance() (*stakemanager.Stakemanager, error) {
	stakeManagerInstance, err := stakemanager.NewStakemanager(GetStakeManagerAddress(), mainChainClient)
	if err != nil {
		Logger.Error("Unable to create stakemanager instance", "error", err)
	}

	return stakeManagerInstance, err
}

func GetStakeManagerABI() (abi.ABI, error) {
	return abi.JSON(strings.NewReader(stakemanager.StakemanagerABI))
}

//
// Get main/matic clients
//

// GetMainClient returns main chain's eth client
func GetMainClient() *ethclient.Client {
	return mainChainClient
}

// GetMaticClient returns matic's eth client
func GetMaticClient() *ethclient.Client {
	return maticClient
}

// GetMaticRPCClient returns matic's RPC client
func GetMaticRPCClient() *rpc.Client {
	return maticRPCClient
}

// GetPrivKey returns priv key object
func GetPrivKey() secp256k1.PrivKeySecp256k1 {
	return privObject
}

// GetPubKey returns pub key object
func GetPubKey() secp256k1.PubKeySecp256k1 {
	return pubObject
}

// GetAddress returns address object
func GetAddress() []byte {
	return GetPubKey().Address().Bytes()
}
