package client_test

import (
	"context"
	"testing"

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
	c, err := client.NewClient()
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
			"alice0",
			[]sdk.Msg{&banktypes.MsgSend{
				FromAddress: suite.client.Accounts["alice0"],
				ToAddress:   suite.client.Accounts["alice1"],
				Amount:      sdk.Coins{sdk.NewInt64Coin("atabi", 100000000000)},
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

func (suite *ClientTestSuite) TestQueryBank() {
	testCases := []struct {
		name string
		req  *banktypes.QueryBalanceRequest
	}{
		{
			name: "alice0 balances",
			req: &banktypes.QueryBalanceRequest{
				Address: suite.client.Accounts["alice0"],
				Denom:   "atabi",
			},
		},
		{
			name: "alice1 balances",
			req: &banktypes.QueryBalanceRequest{
				Address: suite.client.Accounts["alice1"],
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
