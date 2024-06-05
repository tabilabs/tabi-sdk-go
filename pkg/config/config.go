package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Keyring  Keyring   `toml:"keyring"`
	Chain    Chain     `toml:"chain"`
	Tx       Tx        `toml:"tx"`
	Accounts []Account `toml:"accounts"`
}

type Keyring struct {
	Dir     string `toml:"dir"`
	Backend string `toml:"backend"`
}

type Account struct {
	Name         string `toml:"name"`
	Mnemonic     string `toml:"mnemonic"`
	CoinType     uint32 `toml:"coin_type"`
	AccountIndex uint32 `toml:"account_index"`
	AddressIndex uint32 `toml:"address_index"`
}

type Chain struct {
	ChainId  string `toml:"chain_id"`
	GrpcAddr string `toml:"grpc_addr"`
	NodeAddr string `toml:"node_addr"`
}

type Tx struct {
	GasLimit  uint64 `toml:"gas_limit"`
	FeeAmount string `toml:"fee_amount"`
	FeeDenom  string `toml:"fee_denom"`
}

// ClientConfig is the global client config
var (
	ClientConfig Config
)

// ReadConfig reads the config file and sets the ClientConfig
func ReadConfig(path string) {
	if _, err := toml.DecodeFile(path, &ClientConfig); err != nil {
		panic("failed to read config file: " + err.Error())
	}
}
