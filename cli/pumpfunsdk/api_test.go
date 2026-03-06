package pumpfunsdk

import (
	"Zed/mocks"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_GetBondingCurve(t *testing.T) {
	binaryData := []uint8{23, 183, 248, 55, 96, 216, 172, 96, 183, 167, 23, 91, 20, 207, 3, 0, 216, 35, 159, 253, 6, 0, 0, 0, 183, 15, 5, 15, 131, 208, 2, 0, 216, 119, 123, 1, 0, 0, 0, 0, 0, 128, 198, 164, 126, 141, 3, 0, 0}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockIClient(ctrl)

	mockClient.EXPECT().GetAccountInfoWithOpts(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		&rpc.GetAccountInfoResult{
			RPCContext: rpc.RPCContext{},
			Value: &rpc.Account{
				Lamports:   0,
				Owner:      solana.PublicKey{},
				Data:       rpc.DataBytesOrJSONFromBytes(binaryData),
				Executable: false,
				RentEpoch:  nil,
			},
		}, nil)
	bondingCurveKey := solana.MustPublicKeyFromBase58("CmdCbQVR9io8hsK8WxCV8wnrHrqToKudPjamU6HX8zCL")
	b, err := GetBondingCurve(mockClient, bondingCurveKey)
	if err != nil {
		t.Fatal(err)
	}
	if b == nil {
		t.Fail()
	}
	if b.Complete == true {
		t.Fail()
	}
	if b.VirtualTokenReserves != 1072111264704439 {
		t.Fail()
	}
}
