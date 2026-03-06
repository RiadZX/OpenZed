package strategy

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gookit/slog"
	"math/rand"
	"time"
)

type TemporalSend struct {
	apiKey     string
	httpClient IClient
}

var NOZOMI_TIP_ADDRESSES = []string{
	"TEMPaMeCRFAS9EKF53Jd6KpHxgL47uWLcpFArU1Fanq",
	"noz3jAjPiHuBPqiSPkkugaJDkJscPuRhYnSpbi8UvC4",
	"noz3str9KXfpKknefHji8L1mPgimezaiUyCHYMDv1GE",
	"noz6uoYCDijhu1V7cutCpwxNiSovEwLdRHPwmgCGDNo",
	"noz9EPNcT7WH6Sou3sr3GGjHQYVkN3DNirpbvDkv9YJ",
	"nozc5yT15LazbLTFVZzoNZCwjh3yUtW86LoUyqsBu4L",
	"nozFrhfnNGoyqwVuwPAW4aaGqempx4PU6g6D9CJMv7Z",
	"nozievPk7HyK1Rqy1MPJwVQ7qQg2QoJGyP71oeDwbsu",
	"noznbgwYnBLDHu8wcQVCEw6kDrXkPdKkydGJGNXGvL7",
	"nozNVWs5N8mgzuD3qigrCG2UoKxZttxzZ85pvAQVrbP",
	"nozpEGbwx4BcGp6pvEdAh1JoC2CQGZdU6HbNP1v2p6P",
	"nozrhjhkCr3zXT3BiT4WCodYCUFeQvcdUkM7MqhKqge",
	"nozrwQtWhEdrA6W8dkbt9gnUaMs52PdAv5byipnadq3",
	"nozUacTVWub3cL4mJmGCYjKZTnE9RbdY5AP46iQgbPJ",
	"nozWCyTPppJjRuw2fpzDhhWbW355fzosWSzrrMYB1Qk",
	"nozWNju6dY353eMkMqURqwQEoM3SFgEKC6psLCSfUne",
	"nozxNBgWohjR75vdspfxR5H9ceC7XXH99xpxhVGt3Bb",
}

func NewTemporalSend(url string, apiKey string) *TemporalSend {
	return &TemporalSend{
		apiKey:     apiKey,
		httpClient: rpc.New(url),
	}
}

func PickRandomNozomiTipAddress() string {
	rand.NewSource(time.Now().UnixNano())
	return NOZOMI_TIP_ADDRESSES[rand.Intn(len(NOZOMI_TIP_ADDRESSES))]
}

func (t TemporalSend) SendTransaction(transaction solana.Transaction, params SendParams) (success bool, err error) {
	sig, err := t.httpClient.SendTransactionWithOpts(context.Background(), &transaction, rpc.TransactionOpts{SkipPreflight: true, MaxRetries: new(uint)})
	if err != nil {
		slog.Debug("Temporal sending failed: " + err.Error())
		return false, err
	}
	slog.Info(params.TaskName + ": tx sent : " + sig.String())
	return true, nil
}

func (t TemporalSend) GetTipInstructions(sender solana.Wallet, TipFee float64) (instructions []solana.Instruction, err error) {
	wallet := solana.MustPublicKeyFromBase58(PickRandomNozomiTipAddress())
	transferIx := system.NewTransferInstruction(uint64(TipFee*1e9), sender.PublicKey(), wallet).Build()
	instructions = append(instructions, transferIx)
	return instructions, nil
}
