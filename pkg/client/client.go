package client

import "C"
import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	taibapp "github.com/tabilabs/tabi/app"
	tabihd "github.com/tabilabs/tabi/crypto/hd"
	tabiencodec "github.com/tabilabs/tabi/encoding/codec"
	captainstypes "github.com/tabilabs/tabi/x/captains/types"
	claimstypes "github.com/tabilabs/tabi/x/claims/types"
	minttypes "github.com/tabilabs/tabi/x/mint/types"
	tokenconverttypes "github.com/tabilabs/tabi/x/token-convert/types"

	"github.com/tabilabs/tabi-sdk-go/pkg/config"
)

type Client struct {
	Config config.Config

	GRPCConn          *grpc.ClientConn
	RPCClient         rpcclient.Client
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          sdkclient.TxConfig
	Keyring           keyring.Keyring

	// accounts
	Accounts map[string]string

	// tx client
	TxClient tx.ServiceClient

	// query client
	AuthQueryClient authtypes.QueryClient
	BankQueryClient banktypes.QueryClient

	MintQueryClient         minttypes.QueryClient
	CaptainsQueryClient     captainstypes.QueryClient
	TokenConvertQueryClient tokenconverttypes.QueryClient
	ClaimsQueryClient       claimstypes.QueryClient
}

// NewClient creates a new client with config path.
func NewClient(path string) (*Client, error) {
	config.ReadConfig(path)
	return newClient()
}

// newClient creates a new client.
func newClient() (*Client, error) {
	var c Client

	c.Config = config.ClientConfig
	c.Accounts = make(map[string]string)

	c.initEncodingConfig()

	err := c.dial()
	if err != nil {
		return nil, err
	}

	err = c.initKeyring()
	if err != nil {
		return nil, err
	}

	err = c.initAccounts()
	if err != nil {
		return nil, err
	}

	c.initChainClients()

	return &c, nil
}

// initEncodingConfig initializes the encoding config.
func (c *Client) initEncodingConfig() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)

	c.InterfaceRegistry = interfaceRegistry
	c.Codec = codec
	c.TxConfig = authtx.NewTxConfig(codec, authtx.DefaultSignModes)

	tabiencodec.RegisterInterfaces(c.InterfaceRegistry)
	taibapp.ModuleBasics.RegisterInterfaces(c.InterfaceRegistry)
}

// initChainClients initializes the client.
func (c *Client) initChainClients() error {
	rpcClient, err := sdkclient.NewClientFromNode(c.Config.Chain.NodeAddr)
	if err != nil {
		return err
	}

	c.RPCClient = rpcClient
	c.TxClient = tx.NewServiceClient(c.GRPCConn)

	c.AuthQueryClient = authtypes.NewQueryClient(c.GRPCConn)
	c.BankQueryClient = banktypes.NewQueryClient(c.GRPCConn)
	c.MintQueryClient = minttypes.NewQueryClient(c.GRPCConn)
	c.CaptainsQueryClient = captainstypes.NewQueryClient(c.GRPCConn)
	c.TokenConvertQueryClient = tokenconverttypes.NewQueryClient(c.GRPCConn)
	c.ClaimsQueryClient = claimstypes.NewQueryClient(c.GRPCConn)

	return nil
}

// dial dials the gRPC connection.
func (c *Client) dial() error {
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(c.InterfaceRegistry).GRPCCodec()),
		),
	}

	clientConn, err := grpc.Dial(config.ClientConfig.Chain.GrpcAddr, dialOpts...)

	if err != nil {
		return err
	}
	c.GRPCConn = clientConn
	return nil
}

// InitKeyring initializes the Keyring.
func (c *Client) initKeyring() error {
	kr, err := keyring.New("tabi-sdk-go",
		keyring.BackendTest,
		config.ClientConfig.Keyring.Dir,
		strings.NewReader(""),
		c.Codec,
		[]keyring.Option{tabihd.EthSecp256k1Option()}...)
	if err != nil {
		return err
	}
	c.Keyring = kr
	return nil
}

// initAccounts initializes and persists accounts to the Keyring.
func (c *Client) initAccounts() error {
	for _, acc := range config.ClientConfig.Accounts {
		hdPath := hd.CreateHDPath(acc.CoinType, acc.AccountIndex, acc.AddressIndex).String()
		_, err := c.Keyring.NewAccount(acc.Name, acc.Mnemonic, "", hdPath, tabihd.EthSecp256k1)
		if err != nil {
			// we just log error if account already exists
			fmt.Println("warning but never mind: ", err)
		}

		record, err := c.Keyring.Key(acc.Name)
		if err != nil {
			return err
		}

		pubKey, err := record.GetPubKey()
		if err != nil {
			return err
		}

		addr := sdk.AccAddress(pubKey.Address().Bytes()).String()
		c.Accounts[acc.Name] = addr
	}
	return nil
}

// SendTx broadcasts the transaction but accept msgs.
func (c *Client) SendTx(msgs []sdk.Msg, from string) (*tx.BroadcastTxResponse, error) {
	factory := DefaultTxFactory()

	err := factory.RetrieveSeqAndNum(c, c.Accounts[from])
	if err != nil {
		return nil, err
	}

	txBuilder, err := factory.BuildUnsignedTx(c, msgs)
	if err != nil {
		return nil, err
	}

	err = factory.SignTx(c, from, txBuilder)
	if err != nil {
		return nil, err
	}

	txBytes, err := c.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	grpcRes, err := c.TxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return nil, err
	}

	return grpcRes, nil
}

// SendTxWithBlockMode broadcasts the transaction with block mode.
func (c *Client) SendTxWithBlockMode(msgs []sdk.Msg, from string) (*tx.BroadcastTxResponse, error) {
	factory := DefaultTxFactory()

	err := factory.RetrieveSeqAndNum(c, c.Accounts[from])
	if err != nil {
		return nil, err
	}

	txBuilder, err := factory.BuildUnsignedTx(c, msgs)
	if err != nil {
		return nil, err
	}

	err = factory.SignTx(c, from, txBuilder)
	if err != nil {
		return nil, err
	}

	txBytes, err := c.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	grpcRes, err := c.TxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return nil, err
	}

	return grpcRes, nil
}
