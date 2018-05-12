package btc

import (
	"github.com/btcsuite/btcd/rpcclient"
	"sync"
	"blockchain_server/types"
	l4g "github.com/alecthomas/log4go"
	"blockchain_server/chains/btc/btc_settings"
	"blockchain_server/conf"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"errors"
	"sync/atomic"
	"github.com/btcsuite/btcd/btcjson"
	"blockchain_server/utils"
	"bytes"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcd/txscript"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"net/http"
	"fmt"
)

var (
	MaxConfiramtionNumber = 999999
	index_prefix		  = "BTC_HD_Child_PubKey"
	)

type Client struct {
	*rpcclient.Client
	//addresslist []btcutil.Address
	cmMtx       sync.Mutex

	// looks like bitcoin not need this field,
	// the wallet will rescan automatically
	startScanHeight	uint64

	blockHeight		uint64

	// rpc config init from bitcoin settings
	rpc_config *rpcclient.ConnConfig

	blockNotification  chan interface{}
	walletNotification chan interface{}

	quit    chan struct{}
	wg      sync.WaitGroup
	started bool

	quitMtx sync.Mutex

	// 分层确定钱包使用扩散公钥, 和index, 生成子公钥, 并由公钥得到
	// pay-to-pubkey-hash
	// 作为收币地址, 可以说一个index代表一个收币地址,
	// accindexmtx, 用于index, 可能在多协程的情况下的访问同步
	childIndexMtx sync.Mutex

	chain_params *chaincfg.Params

	key_settings *btc_settings.KeySettings
	rpc_settings *btc_settings.RPCSettings

	confirminationNumber uint64
	// the following channel should get from outside call
	// rcTxChannel types.RechargeTxChannel
	rechargeTxNotification types.RechargeTxChannel
}

//func (c *Client)toTx(tx *chain.RelevantTx) error {
//	return nil
//}

func (c *Client) cmlock() {
	c.cmMtx.Lock()
}

func (c *Client) cmunlock() {
	c.cmMtx.Unlock()
}

func (c *Client) WaitForShutdown() {
	c.WaitForShutdown()
	c.wg.Wait()
}

// TODO: to implement client interfaces
func (c *Client)Name() string {
	return types.Chain_bitcoin
}



func (c *Client)SendTx(privkey string, tx *types.Transfer) error {
	var err error
	if tx.From, err = c.virtualKeyToAddress(privkey); err!=nil {
		l4g.Error("Bitcoin cannot deccrypt virtual key stirng to address, error:%s",
			err.Error())
		return err
	}

	if err=c.BuildTx(tx); err!=nil {
		return err
	}
	signedTxBytes, err := c.SignTx(privkey, tx)
	if err!=nil {
		return err
	}
	if err:=c.SendSignedTx(signedTxBytes, tx); err!=nil {
		return err
	}

	return nil
}


func (c *Client)BuildTx(stx *types.Transfer) error {
	var to, change btcutil.Address
	var ads []btcutil.Address
	var err error

	if change, err = btcutil.DecodeAddress(stx.From, c.chain_params); err!=nil {
		return err
	} else { ads = append(ads, change) }

	if to, err = btcutil.DecodeAddress(stx.To, c.chain_params); err!=nil {
		return err
	}

	unspends, err := c.ListUnspentMinMaxAddresses(0, MaxConfiramtionNumber, ads)
	if err!=nil { return err }

	amount, _ := utils.DecimalCvt_i_f(stx.Value, 8, 0).Float64()

	var (
		totalunspent_amount = 0.0
		change_amount       = 0.0
		inputs              []btcjson.TransactionInput
	)

	// 循环utxo, 获取相应资金的utxo
	// var rawTxInput []btcjson.RawTxInput

	//As for formulas, if you use standard addresses (not P2SH), the formula is:
	// fee = (n_inputs * 148 + n_outputs * 34 + 10) * price_per_byte
	//price_perbyte := 5

	// 暂时使用 0.00015/bk 来计算矿工费, 小于1kb, 则提交0.00015
	fee_estimat := 0.00015

	for i, utxo := range unspends {
		//if uint64(utxo.Confirmations) < c.confirminationNumber {
		//	unuseable += utxo.Amount
		//	//continue
		//}

		totalunspent_amount += utxo.Amount
		inputs = append(inputs, btcjson.TransactionInput{Txid: utxo.TxID, Vout: utxo.Vout})

		// maybe should Create []btcjson.RawTxInput and stored some where,
		// as the "[]RawTxinput" param for SignRawTransactions, to sign the transaction
		// rawTxInput = append(rawTxInput,
		// btcjson.RawTxInput{Txid:utxo.TxID, Vout:utxo.Vout, ScriptPubKey:utxo.ScriptPubKey, RedeemScript:utxo.RedeemScript})

		feeincrease := float64(i*180/1000) * 0.00015

		if totalunspent_amount >= (amount + fee_estimat + feeincrease ) {
			fee_estimat += feeincrease
			change_amount = totalunspent_amount - amount - fee_estimat
			break
		}
	}


	if true {
		l4g.Trace("list all unspend utxo!")
		var allutxo float64 =0
		for i, utxo := range unspends {
			l4g.Trace("index:%d, address:%s, amount:%f", i, utxo.Address, utxo.Amount)
			allutxo += utxo.Amount
		}
		l4g.Trace("\nTxInfomation: from:%s, to:%s, need:(%f + %f), have:%f",
			stx.From, stx.To, amount, fee_estimat, allutxo)
	}

	if totalunspent_amount < (amount + fee_estimat) {
		return fmt.Errorf("SendTx Not enougth bitcoin to send! Information:\n" +
			"from:%s, to:%s, need_amount:(%f + %f), have_amount:(%f)",
				stx.From, stx.To, amount, fee_estimat, totalunspent_amount)
	}

	amounts := make(map[btcutil.Address]btcutil.Amount)

	if amounts[to], err = btcutil.NewAmount(amount); err!=nil {
		return err
	}

	// 多余的资金找回给之前的地址
	if change_amount!=0 {
		if amounts[change], err = btcutil.NewAmount(change_amount); err!=nil {
			return nil
		}
	}

	msgTx, err := c.CreateRawTransaction(inputs, amounts, nil)
	if err!=nil { return err }

	for i, txIn := range msgTx.TxIn {
		// 先把ScriptPubkey保存在SignatureScript中,
		// 在SignTx这一步中需要使用
		txIn.SignatureScript, err  = hex.DecodeString(unspends[i].ScriptPubKey)
		if err!=nil { return err }
	}

	txBuf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	if err := msgTx.Serialize(txBuf); err != nil {
		return err
	}

	stx.Fee = uint64(utils.DecimalCvt_f_i(fee_estimat, 0, 8))
	stx.Additional_data = txBuf.Bytes()
	return nil
}

func (c *Client)SignTx(chiperKey string, stx *types.Transfer) ([]byte, error){
	if stx.Additional_data==nil {
		return nil, errors.New("btc sign transaction error, transaciton serialize is nil")
	}

	msgTx := wire.NewMsgTx(wire.TxVersion)
	if err:=msgTx.Deserialize(bytes.NewBuffer(stx.Additional_data)); err!=nil {
		return nil, err
	}

	var (
		index uint32
		err error
		ext_priv *hdkeychain.ExtendedKey
		privKey  *btcec.PrivateKey
	)

	// 根据index找到真正的private key
	index, err = keyToIndex(chiperKey, index_prefix)
	if err!=nil {return nil, err}

	if ext_priv, err = c.key_settings.Ext_pri.Child(index); err!=nil {
		return nil, err
	}
	if privKey, err = ext_priv.ECPrivKey(); err!=nil {
		return nil, err
	}

	// 如果有离线钱包, 可以使用这个方法来签名
	// c.SignRawTransaction2()
	for i, txIn := range msgTx.TxIn {
		txIn.SignatureScript, err = txscript.SignatureScript(msgTx, i, txIn.SignatureScript,
			txscript.SigHashAll, privKey, true)
	}

	txBuf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSize()))
	if err := msgTx.Serialize(txBuf); err != nil {
		return nil, err
	}

	signedTxBuf := txBuf.Bytes()
	stx.Additional_data = signedTxBuf

	return signedTxBuf,nil
}

func (c *Client)SendSignedTx(txByte []byte, tx *types.Transfer) (error) {
	msgTx := wire.NewMsgTx(wire.TxVersion)
	var (
		hash *chainhash.Hash
		err error
	)

	buffer := bytes.NewBuffer(txByte)
	if err = msgTx.Deserialize(buffer); err!=nil {
		return err
	}

	if hash, err = c.SendRawTransaction(msgTx, false);err!=nil {
		l4g.Error("Bitcoin SendRawTransaction error: %s", err.Error())
		return err
	} else {
		tx.Tx_hash = hash.String()
	}

	return nil
}

func (c *Client) toTx(btx *btcjson.GetTransactionResult) (*types.Transfer, error) {
	tx := &types.Transfer{}
	err:=c.updateTxWithBtcTx(tx, btx)

	if err!=nil {
		return nil, err
	}
	return tx, nil
}

func (c *Client) scriptAddes(script []byte)(adds []string, err error) {
	var t *btcjson.DecodeScriptResult
	if t, err = c.DecodeScript(script); err==nil {
		adds = t.Addresses
	}
	return
}

func (c *Client) msgTxOutAddes(txout []*wire.TxOut) (adds[]string, err error) {
	var tmps []string
	for _, txo := range txout {
		if tmps, err = c.scriptAddes(txo.PkScript); err==nil {
			adds = append(adds, tmps[:]...)
		} else {
			l4g.Error("Bitcoin InnerErr:get address(MsgTx.Out), message:%s",
				err.Error())
		}
	}
	return
}

func (c *Client) msgTxFrom(msgTx *wire.MsgTx)(from []string, err error) {
	var tx *btcutil.Tx
	for _, txIn :=  range msgTx.TxIn {
		if tx, err = c.GetRawTransaction(&txIn.PreviousOutPoint.Hash); err==nil {
			var tmps []string
			if tmps, err = c.msgTxOutAddes(tx.MsgTx().TxOut); err!=nil {
				l4g.Error("Bitcoin InnerErr:msgTxAddresses, message:%s",
					err.Error())
			} else {
				index := txIn.PreviousOutPoint.Index
				if uint32(len(tmps))>index {
					from = append(from, tmps[index])
				}
			}
		} else {
			l4g.Error("Bitcoin InnerErr:msgTxAddresses,message %s", err.Error())
		}
	}

	return
}

// 暂时解决思路
// 比特币获取from, to 的思想:
// 首先在遍历txIn, 找出所有from地址
// 遍历detial, 找出不在from中的地址作为to地址(所有txout中的地址, 有可能是找零地址)
// 如果不存在detail, 则解析所有txout中的地址, 用上面的方法来比较, 得出to
// 得出to以后, 在from中, 找到第一个不为to的地址, 作为from,
// to地址对应的txout.Value,就是发送的amount
// 后面需要把所有from和to返回出去.

// blockin, 由于目前这个rpcclient库的transaction没有解析blockin, 所以
// blockin和confirmheight, 暂时都为0
func (c *Client) updateTxWithBtcTx(stx *types.Transfer, btx *btcjson.GetTransactionResult) error {
	//l4g.Trace("Bitcoin Update Transaction: %#v", *btx)
	stx.Tx_hash = btx.TxID
	var (
		serializedTx []byte
		err          error
		value        float64
	)

	serializedTx, err = hex.DecodeString(btx.Hex)
	if err != nil { return err }

	msgTx := &wire.MsgTx{}
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err!=nil { return err }

	froms, err := c.msgTxFrom(msgTx)

	toOk:
	for _, detail := range btx.Details {
		if !utils.StrSilenceContain(froms, detail.Address) {
			stx.To = detail.Address
			value = detail.Amount
			break toOk
		}
	}

	if stx.To=="" {
		for _, txout := range msgTx.TxOut {
			if decodScritp, err := c.DecodeScript(txout.PkScript); len(decodScritp.Addresses)>0 && err!=nil {
				for _, tmp := range decodScritp.Addresses {
					if !utils.StrSilenceContain(froms, tmp) {
						value = btcutil.Amount(txout.Value).ToBTC()
						stx.To = decodScritp.Addresses[0]
					}
				}
			}
		}
	}

	if stx.From=="" {
		for _, tmp := range froms {
			if tmp!= stx.To {
				stx.From = tmp
			}
		}
	}

	if stx.Value==0 {
		stx.Value =  utils.Abs(utils.DecimalCvt_f_i(value, 0, 8))
	}
	if stx.Fee==0 {
		stx.Fee = utils.Abs(utils.DecimalCvt_f_i(btx.Fee, 0, 8))
	}
	if stx.Total==0 {
		stx.Total = stx.Value + stx.Fee
	}

	if uint64(btx.Confirmations) > btc_settings.Client_config().TxConfirmNumber &&
		btx.Confirmations > 0 {
		stx.State = types.Tx_state_confirmed
		//stx.ConfirmatedHeight = stx.InBlock + uint64(btx.Confirmations)
	} else if btx.Confirmations>0 {
		stx.State = types.Tx_state_mined
	} else {
		stx.State = types.Tx_state_commited
	}

	// golang bitcoin rpcclient 'block' have not defined height
	//stx.InBlock = uint64(btx.BlockIndex)

	stx.Time = uint64(btx.Time)
	stx.Token = nil
	stx.Additional_data = nil
	return nil

}
func (c *Client)UpdateTx(stx *types.Transfer) error {
	var(
		hash chainhash.Hash
	)
	if err:=chainhash.Decode(&hash, stx.Tx_hash); err!=nil {
		return err
	}

	btx, err := c.GetTransaction(&hash)
	if err!=nil {
		return err
	}
	return c.updateTxWithBtcTx(stx, btx)
}

func (c *Client)BlockHeight() uint64 {
	return atomic.LoadUint64(&c.blockHeight)
}

// todo : all notifycation use this channel, like the following:
//		select message.(type) {
//		case t1:
//		case t2
func (c *Client)SetNotifyChannel(ch chan interface{}) {
	//c.notification_channel = ch
}

//func (c *Client) loopNotify() {
//	for {
//		c.cmlock()
//		i := len(c. message_queue)
//		msg := c.message_queue[0]
//		c.message_queue = c.message_queue[1:]
//		c.cmunlock()
//
//		select {
//		case c.notification_channel <- msg:
//		case <-c.quit: {
//			l4g.Warn("still have %d message was not deal, and loopNotify will exit!", i)
//			goto out
//		}
//		}
//	}
//	out:
//		c.wg.Done()
//}
//
//func (c *Client)sendNotify(msg interface{}) {
//	c.cmlock()
//	c.message_queue = append(c.message_queue, msg)
//	c.cmunlock()
//}

func (c *Client)SubscribeRechargeTx(txChannel types.RechargeTxChannel) {
	c.rechargeTxNotification = txChannel
}

func (c *Client)InsertRechargeAddress(addresses []string) (invalid []string) {
	labelname := "watchonly"
	for _, v := range addresses {
		if address, err := btcutil.DecodeAddress(v, c.chain_params);err!=nil {
			l4g.Error("bitcoin decode address faild, message:%s", err.Error())
			invalid = append(invalid, v)
		} else {
			//if err := c.SetAccount(address, labelname); err!=nil {
			//	invalid = append(invalid, v)
			//	l4g.Trace("Bitcoin import address error:%s", err.Error())
			//} else {
			//	l4g.Trace("-------::::::::::Bitcoin import address:'%s', to wallet account:'%s'",
			//		address, "receive")
			//}
			if err := c.ImportAddress(address.EncodeAddress(), labelname, false); err!=nil {
				invalid = append(invalid, v)
				l4g.Error("bitcoin import address faild, message:%s", err.Error())
			}
		}
	}
	return
}

func (c *Client) GetBalance(address string, _beNil *string) (uint64, error) {
	if _beNil !=nil {
		l4g.Trace("bitcoin GetBalance,  not support tokens!")
	}
	//c.Client.get
	adds, err := btcutil.DecodeAddress(address, c.chain_params)
	if err!=nil {
		return 0, err
	}

	unspents, err := c.ListUnspentMinMaxAddresses(1, MaxConfiramtionNumber, []btcutil.Address{adds})

	if err!=nil {
		return 0, err
	}

	var balance float64 = 0
	for _, unspent := range unspents {
		balance += unspent.Amount
	}
	return uint64(utils.DecimalCvt_f_i(balance, 1, 8)), nil
}

func (c *Client)Tx(hash string)(*types.Transfer, error) {

	return nil, nil
}

func (c *Client) refresh_blockheight() (uint64, error) {
	if h, err := c.GetBlockCount();  err!=nil {
		return 0, err
	} else {
		atomic.StoreUint64(&c.blockHeight, uint64(h))
		return uint64(h), err
	}
}

func (c *Client) Start() error {
	if !config.IsOnlinemode { return nil }

	// 0 indcates an unlimited number of connection attmpts
	// this is neccessary when a 'ws' client was created with the DisableConnectOnNew
	// field of conn-config struct 'rpcclient.connConfig'
	if c.rpc_config.HTTPPostMode {

		l4g.Trace("bitcoin net connect mode:'http rpc'")

	} else if c.rpc_config.Endpoint=="ws" {

		l4g.Trace("bitcoin net conncet mode:'ws rpc")
		err := c.Connect(0)
		if err != nil { return err }
		// Verify that the server is running on the expected network.
		net, err := c.GetCurrentNet()
		if err != nil {
			c.Disconnect()
			return err
		}
		if net != c.chain_params.Net {
			c.Disconnect()
			return errors.New("mismatched networks")
		}
	}

	// 发送一个close 到服务器, 服务器收到后退出循环
	http.Post("http://127.0.0.1" + c.rpc_settings.Http_server + "/isok",
		"text/plain; charset=utf-8", bytes.NewBuffer([]byte(`close`)))

	c.quitMtx.Lock()
	go c.startHttpServer()
	go c.handler()
	c.started = true
	c.wg.Add(2)
	c.quitMtx.Unlock()
	return nil
}

// Stop disconnects the client and signals the shutdown of all goroutines
// started by Start.
func (c *Client) Stop() {
	c.quitMtx.Lock()
	select {
	case <-c.quit:
	default:
		close(c.quit)
		c.Shutdown()

		if !c.started {
			close(c.walletNotification)
			close(c.blockNotification)
		}
	}
	c.quitMtx.Unlock()
}

