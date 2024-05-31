# Tabi-sdk-go

Tabi-sdk-go is a Go library designed for building applications that interact with the Tabi blockchain. Currently, it serves as an internal tool for Tabi Labs.

It primarily implements a simplified version of the Cosmos SDK `TxFactory`, which is used to build and sign transactions. Note that it only supports `SignMode_SIGN_MODE_DIRECT`. 

The SDK uses a keyring to manage your keys, which are stored in the `keyring-dir` directory. Please handle these keys with care.

## Quick Start

Configure the SDK client using the `./config/config.toml` file, which is the default path for the client configuration.

A template file is provided for convenience. You can copy it to your project and modify it as needed.

```toml
[chain]
chain_id = "tabi_9789-1"
grpc_addr = "localhost:9090"

[tx]
gas_limit = 4000000
fee_amount = "8000000000000000"
fee_denom = "atabi"

[keyring]
dir = "SET YOUR KEYRING DIR"
backend = "test"

[[accounts]]
name = "alice0"
mnemonic = "SET YOUR MNEMONIC"
coin_type = 60
account_index = 0
address_index = 0

[[accounts]]
name = "alice1"
mnemonic = "SET YOUR MNEMONIC"
coin_type = 60
account_index = 1
address_index = 0
```

Create a new client instance with `NewClient()` or `NewClientWithConfigPath()` and use it to send transactions.

Note that `NewClient()` use the default configuration file path `./config/config.toml` while `NewClientWithConfigPath()` allows you to specify a custom path.

```go
package main

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tabilabs/tabi-sdk-go/client"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func main() {
	client := client.NewClient()

	msgs := []sdk.Msg{
		&types.MsgSend{
			FromAddress: client.Accounts["alice0"].Address,
			ToAddress:   client.Accounts["alice1"].Address,
			Amount:      sdk.Coins{sdk.NewInt64Coin("atabi", 100000000000)},
		},
	}

	resp, err := client.SendTx(msgs, "alice0")

	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Response:", resp)
	}
}
```

Query functionality is also supported. However, you may need to import some query clients if they are not already included.

```go
type Client struct {
	// Existing query clients
	AuthQueryClient authtypes.QueryClient
	BankQueryClient banktypes.QueryClient

	// Add more clients here as needed
	MintQueryClient         minttypes.QueryClient
	CaptainsQueryClient     captainstypes.QueryClient
	TokenConvertQueryClient tokenconverttypes.QueryClient
	ClaimsQueryClient       claimstypes.QueryClient
}
```