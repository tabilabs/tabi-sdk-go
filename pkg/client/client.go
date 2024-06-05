package client

import "C"
import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	taibapp "github.com/tabilabs/tabi/app"
	tabihd "github.com/tabilabs/tabi/crypto/hd"
	tabiencoding "github.com/tabilabs/tabi/encoding"
	captainstypes "github.com/tabilabs/tabi/x/captains/types"
	claimstypes "github.com/tabilabs/tabi/x/claims/types"
	minttypes "github.com/tabilabs/tabi/x/mint/types"
	tokenconverttypes "github.com/tabilabs/tabi/x/token-convert/types"

	"github.com/tabilabs/tabi-sdk-go/pkg/config"
)

type Client struct {
	clientcfg config.Config

	conn *grpc.ClientConn

	EncodingConfig params.EncodingConfig
	Keyring        keyring.Keyring

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

func newClient() (*Client, error) {
	var c Client

	c.clientcfg = config.ClientConfig
	c.EncodingConfig = tabiencoding.MakeConfig(taibapp.ModuleBasics)
	c.Accounts = make(map[string]string)

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

func (c *Client) dial() error {
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.ForceCodec(codec.NewProtoCodec(c.EncodingConfig.InterfaceRegistry).GRPCCodec()),
		),
	}

	clientConn, err := grpc.Dial(config.ClientConfig.Chain.GrpcAddr, dialOpts...)

	if err != nil {
		return err
	}
	c.conn = clientConn
	return nil
}

// Init initializes the client.
func (c *Client) initChainClients() {
	c.TxClient = tx.NewServiceClient(c.conn)
	c.AuthQueryClient = authtypes.NewQueryClient(c.conn)
	c.BankQueryClient = banktypes.NewQueryClient(c.conn)
	c.MintQueryClient = minttypes.NewQueryClient(c.conn)
	c.CaptainsQueryClient = captainstypes.NewQueryClient(c.conn)
	c.TokenConvertQueryClient = tokenconverttypes.NewQueryClient(c.conn)
	c.ClaimsQueryClient = claimstypes.NewQueryClient(c.conn)
}

// InitKeyring initializes the Keyring.
func (c *Client) initKeyring() error {
	kr, err := keyring.New("tabi-sdk-go",
		keyring.BackendTest,
		config.ClientConfig.Keyring.Dir,
		strings.NewReader(""),
		c.EncodingConfig.Codec,
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

		pubkey, err := record.GetPubKey()
		if err != nil {
			return err
		}

		addr := sdk.AccAddress(pubkey.Address().Bytes()).String()
		c.Accounts[acc.Name] = addr
	}
	return nil
}

// SendTx broadcasts the transaction but accept msgs.
func (c *Client) SendTx(msgs []sdk.Msg, from string) (string, error) {
	factory := DefaultTxFactory()

	err := factory.RetrieveSeqAndNum(c, c.Accounts[from])
	if err != nil {
		return "", err
	}

	txBuilder, err := factory.BuildUnsignedTx(c, msgs)
	if err != nil {
		return "", err
	}

	err = factory.SignTx(c, from, txBuilder)
	if err != nil {
		return "", err
	}

	txBytes, err := c.EncodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", err
	}

	grpcRes, err := c.TxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return "", err
	}

	return grpcRes.String(), nil
}
