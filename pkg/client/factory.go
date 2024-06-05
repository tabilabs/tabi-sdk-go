package client

import (
	"context"
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tabilabs/tabi-sdk-go/pkg/config"
)

// TxFactory is a simplified version of sdk.TxFactory, supporting only direct signing mode.
type TxFactory struct {
	gas     uint64
	chainID string
	fees    sdk.Coins

	accountNumber uint64
	sequence      uint64

	signMode signing.SignMode
}

func DefaultTxFactory() *TxFactory {
	amount := sdk.MustNewDecFromStr(config.ClientConfig.Tx.FeeAmount).RoundInt()
	fees := sdk.NewCoin(config.ClientConfig.Tx.FeeDenom, amount)

	return &TxFactory{
		accountNumber: 0,
		sequence:      0,
		gas:           config.ClientConfig.Tx.GasLimit,
		chainID:       config.ClientConfig.Chain.ChainId,
		fees:          sdk.NewCoins(fees),
		signMode:      signing.SignMode_SIGN_MODE_DIRECT,
	}
}

// RetrieveSeqAndNum retrieves accountNumber and sequence from chain
func (f *TxFactory) RetrieveSeqAndNum(c *Client, addr string) error {
	resp, err := c.AuthQueryClient.Account(context.Background(),
		&authtypes.QueryAccountRequest{Address: addr})
	if err != nil {
		return err
	}

	// unpack and set num and seq
	var acc authtypes.AccountI
	if err := c.InterfaceRegistry.UnpackAny(resp.Account, &acc); err != nil {
		return err
	}

	f.accountNumber = acc.GetAccountNumber()
	f.sequence = acc.GetSequence()

	return nil
}

// BuildUnsignedTx creates a new unsigned transaction
func (f *TxFactory) BuildUnsignedTx(c *Client, msgs []sdk.Msg) (client.TxBuilder, error) {
	tx := c.TxConfig.NewTxBuilder()
	if err := tx.SetMsgs(msgs...); err != nil {
		return nil, err
	}

	tx.SetFeeAmount(f.fees)
	tx.SetGasLimit(f.gas)
	return tx, nil
}

// SignTx signs a transaction using the provided name
func (f *TxFactory) SignTx(c *Client, name string, txBuilder client.TxBuilder) error {
	if c.Keyring == nil {
		return errors.New("keyring must be set prior to signing a transaction")
	}

	k, err := c.Keyring.Key(name)
	if err != nil {
		return err
	}

	pubKey, err := k.GetPubKey()
	if err != nil {
		return err
	}

	signerData := authsigning.SignerData{
		ChainID:       f.chainID,
		AccountNumber: f.accountNumber,
		Sequence:      f.sequence,
		PubKey:        pubKey,
		Address:       sdk.AccAddress(pubKey.Address()).String(),
	}

	sig := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  f.signMode,
			Signature: nil,
		},
		Sequence: f.sequence,
	}

	sigs := []signing.SignatureV2{sig}
	if err := txBuilder.SetSignatures(sigs...); err != nil {
		return err
	}

	bytesToSign, err := c.TxConfig.SignModeHandler().GetSignBytes(f.signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return err
	}

	// sign tx bytes
	sigBytes, _, err := c.Keyring.Sign(name, bytesToSign)
	if err != nil {
		return err
	}

	// construct signature
	sig = signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  f.signMode,
			Signature: sigBytes,
		},
		Sequence: f.sequence,
	}

	return txBuilder.SetSignatures(sig)
}
