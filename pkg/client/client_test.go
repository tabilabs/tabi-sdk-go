package client_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	captainstypes "github.com/tabilabs/tabi/x/captains/types"

	"github.com/tabilabs/tabi-sdk-go/pkg/client"
)

type ClientTestSuite struct {
	suite.Suite

	client *client.Client
}

var s *ClientTestSuite

func TestSdkClientTestSuite(t *testing.T) {
	s = new(ClientTestSuite)
	suite.Run(t, s)
}

func (suite *ClientTestSuite) SetupTest() {
	c, err := client.NewClient("../../config/local.toml")
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.client = c
}

func (suite *ClientTestSuite) TestSendTx() {
	testCases := []struct {
		name string
		from string
		msgs []sdk.Msg
	}{
		{
			"bank send",
			"node0",
			[]sdk.Msg{&banktypes.MsgSend{
				FromAddress: suite.client.Accounts["node0"],
				ToAddress:   suite.client.Accounts["committer"],
				Amount:      sdk.Coins{sdk.NewCoin("atabi", sdk.MustNewDecFromStr("10000000000000000").RoundInt())},
			},
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.client.SendTx(tc.msgs, tc.from)
			suite.Require().NoError(err)
			suite.T().Logf("Response: %v", resp)
		})
	}
}

func (suite *ClientTestSuite) TestGetTx() {
	testCases := []struct {
		name   string
		txHash string
	}{
		{
			"get tx",
			"F2C06FDB1EDB4CC9E335E16048A032C94070B019EB58436EC6B4A76CEE7031C3",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.client.TxClient.GetTx(context.Background(), &tx.GetTxRequest{Hash: tc.txHash})
			suite.Require().NoError(err)
			suite.Require().Equal(uint32(0), resp.TxResponse.Code)
			//suite.T().Logf("TxResponse: %v", resp.TxResponse)

			resp2, err := suite.client.RPCClient.Tx(context.Background(), []byte(tc.txHash), false)
			suite.Require().NoError(err)
			suite.T().Logf("Response: %v", resp2)
		})
	}
}

func (suite *ClientTestSuite) TestQueryBalances() {
	testCases := []struct {
		name string
		req  *banktypes.QueryBalanceRequest
	}{
		{
			name: "committer balances",
			req: &banktypes.QueryBalanceRequest{
				Address: suite.client.Accounts["committer"],
				Denom:   "atabi",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.client.BankQueryClient.Balance(context.Background(), tc.req)
			suite.Require().NoError(err)
			suite.T().Logf("Response: %v", resp)
		})
	}
}

func (suite *ClientTestSuite) TestQueryCaptains() {
	testCases := []struct {
		name string
		req  *captainstypes.QueryCurrentEpochRequest
	}{
		{
			name: "current epoch",
			req:  &captainstypes.QueryCurrentEpochRequest{},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.client.CaptainsQueryClient.CurrentEpoch(context.Background(), tc.req)
			suite.Require().NoError(err)
			suite.T().Logf("Response: %v", resp)
		})
	}
}
