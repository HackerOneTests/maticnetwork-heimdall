package pier

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/streadway/amqp"
	"github.com/tendermint/tendermint/libs/log"

	authTypes "github.com/maticnetwork/heimdall/auth/types"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

const (
	connector = "queue-connector"

	// exchanges
	broadcastExchange = "bridge.exchange.broadcast"
	// heimdall queue
	heimdallBroadcastQueue = "bridge.queue.heimdall"
	// bor queue
	borBroadcastQueue = "bridge.queue.bor"

	// heimdall routing key
	heimdallBroadcastRoute = "bridge.route.heimdall"
	// bor routing key
	borBroadcastRoute = "bridge.route.bor"
)

// QueueConnector queue connector
type QueueConnector struct {
	// URL for connecting to AMQP
	connection *amqp.Connection
	// create a channel
	channel *amqp.Channel
	// tx encoder
	cliCtx cliContext.CLIContext
	// logger
	logger log.Logger
}

// NewQueueConnector creates a connector object which can be used to connect/send/consume bytes from queue
func NewQueueConnector(cdc *codec.Codec, dialer string) *QueueConnector {
	cliCtx := cliContext.NewCLIContext().WithCodec(cdc)
	cliCtx.BroadcastMode = client.BroadcastAsync
	cliCtx.TrustNode = true

	// amqp dialer
	conn, err := amqp.Dial(dialer)
	if err != nil {
		panic(err)
	}

	// initialize exchange
	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	// queue connector
	connector := QueueConnector{
		connection: conn,
		channel:    channel,
		cliCtx:     cliCtx,
		// create logger
		logger: Logger.With("module", "queue-connector"),
	}

	// connector
	return &connector
}

// Start connector
func (qc *QueueConnector) Start() error {
	// exchange declare
	if err := qc.channel.ExchangeDeclare(
		broadcastExchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	); err != nil {
		return err
	}

	// Heimdall

	// queue declare
	if _, err := qc.channel.QueueDeclare(
		heimdallBroadcastQueue, // name
		true,                   // durable
		false,                  // delete when usused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
	); err != nil {
		return err
	}

	// bind queue
	if err := qc.channel.QueueBind(
		heimdallBroadcastQueue, // queue name
		heimdallBroadcastRoute, // routing key
		broadcastExchange,      // exchange
		false,
		nil,
	); err != nil {
		return err
	}

	// start consuming
	msgs, err := qc.channel.Consume(
		heimdallBroadcastQueue, // queue
		heimdallBroadcastQueue, // consumer  -- consumer identifier
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return err
	}

	// process heimdall broadcast messages
	go qc.handleHeimdallBroadcastMsgs(msgs)

	// Bor

	// queue declare
	if _, err := qc.channel.QueueDeclare(
		borBroadcastQueue, // name
		true,              // durable
		false,             // delete when usused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	); err != nil {
		return err
	}

	// bind queue
	if err := qc.channel.QueueBind(
		borBroadcastQueue, // queue name
		borBroadcastRoute, // routing key
		broadcastExchange, // exchange
		false,
		nil,
	); err != nil {
		return err
	}

	// start consuming
	msgs, err = qc.channel.Consume(
		borBroadcastQueue, // queue
		borBroadcastQueue, // consumer  -- consumer identifier
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return err
	}

	// process bor broadcast messages
	go qc.handleBorBroadcastMsgs(msgs)

	return nil
}

// Stop connector
func (qc *QueueConnector) Stop() {
	// close channel & connection
	qc.channel.Close()
	qc.connection.Close()
}

//
// Publish
//

// BroadcastToHeimdall broadcasts to heimdall
func (qc *QueueConnector) BroadcastToHeimdall(msg sdk.Msg) error {
	data, err := qc.cliCtx.Codec.MarshalJSON(msg)
	if err != nil {
		return err
	}

	return qc.BroadcastBytesToHeimdall(data)
}

// BroadcastBytesToHeimdall broadcasts bytes to heimdall
func (qc *QueueConnector) BroadcastBytesToHeimdall(data []byte) error {
	if err := qc.channel.Publish(
		broadcastExchange,      // exchange
		heimdallBroadcastRoute, // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		}); err != nil {
		return err
	}

	return nil
}

// BroadcastToBor broadcasts to bor
func (qc *QueueConnector) BroadcastToBor(data []byte) error {
	if err := qc.channel.Publish(
		broadcastExchange, // exchange
		borBroadcastRoute, // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		}); err != nil {
		return err
	}

	return nil
}

//
// Consume
//

func (qc *QueueConnector) handleHeimdallBroadcastMsgs(amqpMsgs <-chan amqp.Delivery) {
	// tx encoder
	txEncoder := helper.GetTxEncoder()
	// chain id
	chainID := helper.GetGenesisDoc().ChainID
	// current address
	address := hmTypes.BytesToHeimdallAddress(helper.GetAddress())
	// fetch from APIs
	var account authTypes.Account
	response, _ := FetchFromAPI(qc.cliCtx, fmt.Sprintf(GetHeimdallServerEndpoint(AccountDetailsURL), address))
	// get proposer from response
	if err := qc.cliCtx.Codec.UnmarshalJSON(response.Result, &account); err != nil {
		panic(err)
	}

	// get account number and sequence
	accNum := account.GetAccountNumber()
	accSeq := account.GetSequence()

	for amqpMsg := range amqpMsgs {
		var msg sdk.Msg
		if err := qc.cliCtx.Codec.UnmarshalJSON(amqpMsg.Body, &msg); err != nil {
			amqpMsg.Reject(false)
			return
		}

		txBldr := authTypes.NewTxBuilderFromCLI().
			WithTxEncoder(txEncoder).
			WithAccountNumber(accNum).
			WithSequence(accSeq).
			WithChainID(chainID)
		if _, err := helper.BuildAndBroadcastMsgs(qc.cliCtx, txBldr, []sdk.Msg{msg}); err != nil {
			amqpMsg.Reject(false)
			return
		}

		// send ack
		amqpMsg.Ack(true)

		// increment account sequence
		accSeq = accSeq + 1
	}
}

func (qc *QueueConnector) handleBorBroadcastMsgs(amqpMsgs <-chan amqp.Delivery) {
	maticClient := helper.GetMaticClient()

	for amqpMsg := range amqpMsgs {
		var msg ethereum.CallMsg
		if err := json.Unmarshal(amqpMsg.Body, &msg); err != nil {
			amqpMsg.Ack(false)
			qc.logger.Error("Error while parsing the transaction from queue", "error", err)
			return
		}

		// get auth
		auth, err := helper.GenerateAuthObj(maticClient, msg)
		if err != nil {
			amqpMsg.Ack(false)
			qc.logger.Error("Error while fetching the transaction param details", "error", err)
			return
		}

		// Create the transaction, sign it and schedule it for execution
		rawTx := types.NewTransaction(auth.Nonce.Uint64(), *msg.To, msg.Value, auth.GasLimit, auth.GasPrice, msg.Data)

		// signer
		signedTx, err := auth.Signer(types.HomesteadSigner{}, auth.From, rawTx)
		if err != nil {
			amqpMsg.Ack(false)
			qc.logger.Error("Error while signing the transaction", "error", err)
			return
		}

		// broadcast transaction
		if err := maticClient.SendTransaction(context.Background(), signedTx); err != nil {
			amqpMsg.Ack(false)
			qc.logger.Error("Error while broadcasting the transaction", "error", err)
			return
		}

		// send ack
		amqpMsg.Ack(false)
	}
}

// // ConsumeHeimdallQ consumes messages from heimdall queue
// func (qc *QueueConnector) ConsumeHeimdallQ() error {
// 	conn, err := amqp.Dial(qc.AmqpDailer)
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close()

// 	q, err := ch.QueueDeclare(
// 		qc.HeimdallQueue, // name
// 		false,            // durable
// 		false,            // delete when unused
// 		false,            // exclusive
// 		false,            // no-wait
// 		nil,              // arguments
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	msgs, err := ch.Consume(
// 		q.Name, // queue
// 		"",     // consumer
// 		true,   // auto-ack
// 		false,  // exclusive
// 		false,  // no-local
// 		false,  // no-wait
// 		nil,    // args
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	forever := make(chan bool)

// 	go func() {
// 		for d := range msgs {
// 			qc.Logger.Debug("Sending transaction to heimdall", "TxBytes", d.Body)
// 			resp, err := helper.BroadcastTxBytes(qc.cliCtx, d.Body, "")
// 			if err != nil {
// 				qc.Logger.Error("Unable to send transaction to heimdall", "error", err)
// 			} else {
// 				qc.Logger.Info("Sent to heimdall", "Response", resp.String())
// 				// TODO identify msg type checkpoint and add conditional
// 				qc.DispatchToEth(resp.TxHash)
// 			}
// 		}
// 	}()
// 	qc.Logger.Info("Starting queue consumer")
// 	<-forever
// 	return nil
// }

// // ConsumeCheckpointQ consumes checkpoint tx hash from heimdall and sends prevotes to contract
// func (qc *QueueConnector) ConsumeCheckpointQ() error {
// 	// On confirmation/rejection for tx
// 	// Send checkpoint to rootchain incase
// 	conn, err := amqp.Dial(qc.AmqpDailer)
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()
// 	ch, err := conn.Channel()
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close()

// 	q, err := ch.QueueDeclare(
// 		qc.CheckpointQueue, // name
// 		false,              // durable
// 		false,              // delete when unused
// 		false,              // exclusive
// 		false,              // no-wait
// 		nil,                // arguments
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	msgs, err := ch.Consume(
// 		q.Name, // queue
// 		"",     // consumer
// 		true,   // auto-ack
// 		false,  // exclusive
// 		false,  // no-local
// 		false,  // no-wait
// 		nil,    // args
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	forever := make(chan bool)

// 	go func() {
// 		for d := range msgs {
// 			qc.Logger.Debug("Sending transaction to heimdall", "TxBytes", d.Body)
// 			// resp, err := helper.SendTendermintRequest(qc.cliContext, d.Body, helper.BroadcastAsync)
// 			// if err != nil {
// 			// 	qc.Logger.Error("Unable to send transaction to heimdall", "error", err)
// 			// } else {
// 			// 	qc.Logger.Info("Sent to heimdall", "Response", resp.String())
// 			// 	// TODO identify msg type checkpoint and add conditional
// 			// 	qc.DispatchToEth(resp.TxHash)
// 			// }
// 		}
// 	}()
// 	qc.Logger.Info("Starting queue consumer")
// 	<-forever
// 	return nil

// }
