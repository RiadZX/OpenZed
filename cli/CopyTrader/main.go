package copytrader

import (
	"Zed/db"
	"Zed/models"
	"Zed/triggers"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/gookit/slog"
	"github.com/weeaa/goyser"
	geyser_pb "github.com/weeaa/goyser/pb"
	"google.golang.org/grpc/metadata"
)

var (
	recentBlockHash solana.Hash
	tokenProgram    = solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
)

func LaunchCopyTrader(ctx context.Context, userConfig *models.Config, programs []CopyTrader, grpc bool, db *db.Connection) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if userConfig == nil {
		return fmt.Errorf("userConfig is nil")
	}
	if len(programs) == 0 {
		return fmt.Errorf("no programs to launch")
	}
	slog.Info("All tasks count: " + fmt.Sprint(len(userConfig.CopyTraderConfig.AllTasks)))
	if userConfig.CopyTraderConfig.ActiveTasks == nil || len(userConfig.CopyTraderConfig.ActiveTasks) == 0 {
		slog.Error("No tasks to run")
		return nil
	}
	slog.Info("Active tasks count: " + fmt.Sprint(len(userConfig.CopyTraderConfig.ActiveTasks)))
	for p := range programs {
		programs[p].SetDatabase(db)
	}

	err := InitPrograms(programs)
	if err != nil {
		return err
	}

	err = StartPrograms(programs)
	if err != nil {
		return err
	}

	// Start the websocket or grpc connection
	if !grpc {
		//remove the pumpfun_amm program from the list of programs if it was the last one
		if programs[len(programs)-1].GetProgramName() == "PumpFun AMM" {
			programs = programs[:len(programs)-1]
		}
		err = StartWebsocket(ctx, userConfig, programs)
		if err != nil {
			slog.Error(err.Error())
		}
	} else {
		err = StartGRPC(ctx, userConfig, programs)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	return nil
}

// InitPrograms initialize the active copy trading programs
func InitPrograms(programs []CopyTrader) error {
	//Init each program
	for _, program := range programs {
		err := program.Init()
		if err != nil {
			return err
		}
	}
	slog.Debug("Programs initialized")
	return nil
}

// StartPrograms calls Start on the active programs
func StartPrograms(programs []CopyTrader) error {
	//Start each program
	for _, program := range programs {
		err := program.Start()
		if err != nil {
			return err
		}
	}
	for _, program := range programs {
		slog.Info("Program Started : " + program.GetProgramName())
	}
	return nil
}

func StartGRPC(ctxParent context.Context, userConfig *models.Config, programs []CopyTrader) error {
	if ctxParent == nil {
		return fmt.Errorf("context is nil")
	}
	slog.Info("Starting grpc connection")
	err := triggers.NewMonitor(userConfig, userConfig.PnlDelayMS, triggers.RPCOpts{
		Client:  rpc.New(userConfig.RPCUrl),
		Retries: int(userConfig.CopyTraderConfig.GetAccountInfoRetries),
		DelayMS: userConfig.CopyTraderConfig.GetAccountInfoDelayMS,
	})
	if err != nil {
		slog.Error("Error creating monitor: %v", err)
		return err
	}
	for {

		ctx, cancel := context.WithCancel(ctxParent)
		go startBlockHashRefresher(ctx, userConfig.HttpClient, rpc.CommitmentFinalized)
		for {
			//wait until the blockhash is set
			if recentBlockHash.String() != "" {
				break
			}
			time.Sleep(1 * time.Second)
		}
		select {
		case <-ctxParent.Done():
			cancel()
			return fmt.Errorf("parent context cancelled")
		default:
			slog.Debug("BlockHash refresher started")

			geyserRPC := userConfig.GRPCUrl
			md := metadata.MD{}
			if userConfig.GRPCHeaders != "" {
				headers := strings.Split(userConfig.GRPCHeaders, ",")
				for _, header := range headers {
					headerSplit := strings.Split(header, ":")
					if len(headerSplit) == 2 {
						md.Append(headerSplit[0], headerSplit[1])
					}
				}
			}
			slog.Debug("gRPC Metadata:")
			for k, v := range md {
				slog.Debug(k + ":" + strings.Join(v, ","))
			}
			client, err := goyser.New(ctx, geyserRPC, md)
			if err != nil {
				slog.Error("Error creating Geyser client: %v", err)
				cancel()
				time.Sleep(time.Second * 2) // Simple backoff, consider exponential backoff
				continue
			}

			if err = client.AddStreamClient(ctx, "main"); err != nil {
				slog.Error("Error creating subscribe client: %v", err)
				cancel()
				time.Sleep(time.Second * 2) // Simple backoff, consider exponential backoff
				continue
			}

			streamClient := client.GetStreamClient("main")
			if streamClient == nil {
				slog.Error("Client does not have a stream named main")
				cancel()
				time.Sleep(time.Second * 2) // Simple backoff, consider exponential backoff
				continue
			}

			var accountsExclude []string
			var accountsInclude []string

			//add the copied wallet address
			for _, task := range userConfig.CopyTraderConfig.ActiveTasks {
				if !slices.Contains(accountsInclude, task.WalletToCopy) {
					slog.Debug("Adding wallet to copy: " + task.WalletToCopy)
					accountsInclude = append(accountsInclude, task.WalletToCopy)
				}
				//add my wallets
				if !slices.Contains(accountsInclude, task.Wallet.PublicKey().String()) {
					slog.Debug("Adding my wallet: " + task.Wallet.PublicKey().String())
					accountsInclude = append(accountsInclude, task.Wallet.PublicKey().String())
				}
			}

			if err = streamClient.SubscribeTransaction("transactions", &geyser_pb.SubscribeRequestFilterTransactions{
				Vote:            new(bool),
				Failed:          new(bool),
				Signature:       nil,
				AccountInclude:  accountsInclude,
				AccountExclude:  accountsExclude,
				AccountRequired: nil,
			}); err != nil {
				slog.Error("Error subscribing to transactions: %v", err)
				cancel()
				time.Sleep(time.Second * 2) // Simple backoff, consider exponential backoff
				continue
			}
			slog.Info("Subscribed to transactions")
			go func() {
				for {
					select {
					case out, ok := <-streamClient.ErrCh:
						if !ok {
							slog.Error("Stream error channel closed")
							cancel()
							return
						}
						slog.Error("Stream error: " + out.Error())
						cancel()
						return
					case <-ctx.Done():
						return
					}
				}
			}()

			// Send geyser_pg.PingRequest every 15 seconds
			go func() {
				var i int32 = 0
				for {
					select {
					case <-ctx.Done():
						return
					default:
						i++
						_, err := client.Ping(i)
						slog.Info("Pinging...")
						if err != nil {
							if err.Error() == "EOF" {
								slog.Error("EOF error, reconnecting...")
								cancel()
								return
							} else {
								slog.Error("Error pinging: " + err.Error())
								cancel()
								return
							}
						}
						time.Sleep(15 * time.Second)
					}
				}
			}()
			if userConfig.PnlDelayMS > 0 {
				triggers.MonitorInstance.Start()
			}
			go func() {
				for {
					for out := range streamClient.Ch {
						//Check filter name
						for _, program := range programs {
							go program.HandleGRPCMessage(out.GetTransaction(), &recentBlockHash)
						}
					}
				}
			}()

			<-ctx.Done()
			slog.Info("Attempting to reconnect...")
			time.Sleep(time.Second * 1) // Simple backoff, consider exponential backoff
		}
	}
}

func StartWebsocket(ctxParent context.Context, userConfig *models.Config, programs []CopyTrader) error {
	if ctxParent == nil {
		return fmt.Errorf("context is nil")
	}
	if userConfig.HttpClient == nil {
		return fmt.Errorf("http client is nil")
	}
	slog.Info("Starting websocket connection")
	for {
		ctx, cancel := context.WithCancel(ctxParent)
		go startBlockHashRefresher(ctx, userConfig.HttpClient, rpc.CommitmentFinalized)
		for {
			//wait until the blockhash is set
			if recentBlockHash.String() != "" {
				break
			}
			time.Sleep(1 * time.Second)
		}
		select {
		case <-ctxParent.Done():
			cancel()
			return fmt.Errorf("parent context cancelled")
		default:
			slog.Debug("BlockHash refresher started")

			client, err := ws.Connect(ctxParent, userConfig.WSSUrl)
			if err != nil {
				slog.Error("Error connecting to websocket: %v" + err.Error())
				cancel()                    //cancels the blockhash refresher
				time.Sleep(time.Second * 2) // Simple backoff, consider exponential backoff
				continue
			}

			listenProgram := tokenProgram
			isPumpFunRunning := false
			isRaydiumRunning := false
			//if only pumpfun is running, listen to pumpfun logs only
			//if only raydium is running, listen to raydium logs only
			//if both are running, listen to tokenProgram
			for _, program := range programs {
				if program.GetProgramName() == "PumpFun" {
					isPumpFunRunning = true
				}
				if program.GetProgramName() == "Raydium" {
					isRaydiumRunning = true
				}
			}
			if isPumpFunRunning && !isRaydiumRunning {
				listenProgram = solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P")
			}
			if !isPumpFunRunning && isRaydiumRunning {
				listenProgram = solana.MustPublicKeyFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")
			}

			sub, err := client.LogsSubscribeMentions(listenProgram, rpc.CommitmentConfirmed)
			if err != nil {
				slog.Error("Error subscribing to logs: %v" + err.Error())
				client.Close()
				cancel()                    //cancels the blockhash refresher
				time.Sleep(time.Second * 2) // Simple backoff, consider exponential backoff
				continue
			}

			for {
				got, err := sub.Recv()
				if err != nil {
					slog.Error("Error receiving message: " + err.Error())
					sub.Unsubscribe()
					client.Close()
					cancel() //cancels the blockhash refresher
					break    // Break out of the loop to attempt reconnection
				}

				for _, program := range programs {
					go program.HandleWSMessage(got, &recentBlockHash)
				}
			}

			slog.Info("Attempting to reconnect...")
			time.Sleep(time.Second * 2) // Simple backoff, consider exponential backoff
		}
	}
}

// startBlockHashRefresher starts a goroutine that refreshes the blockhash every 30 seconds
func startBlockHashRefresher(ctx context.Context, client models.IClient, commitmentType rpc.CommitmentType) {
	if client == nil {
		slog.Error("Client is nil")
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				out, err := client.GetLatestBlockhash(ctx, commitmentType)
				if err != nil {
					slog.Info("1) Error getting recent blockhash: " + err.Error())
					time.Sleep(5 * time.Second)
					continue
				}
				recentBlockHash = out.Value.Blockhash
				slog.Debug("Got recent blockhash: " + recentBlockHash.String())
				time.Sleep(15 * time.Second)
			}
		}
	}()
	<-ctx.Done()
}
