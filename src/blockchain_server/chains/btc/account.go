package btc

import (
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"blockchain_server/types"
	"blockchain_server/conf"
	"fmt"
	l4g "github.com/alecthomas/log4go"
)

func (c *Client) virtualKeyToRealPubkey(privkey string) (*hdkeychain.ExtendedKey, error) {
	if c.key_settings==nil || c.key_settings.Ext_pub==nil {
		return nil, fmt.Errorf("KeySettings invalid!")
	}

	index, err := keyToIndex(privkey, index_prefix)
	if err!=nil {
		return nil, err
	}

	return c.key_settings.Ext_pub.Child(index)
}

func (c *Client) virtualKeyToRealPrikey(privkey string) (*hdkeychain.ExtendedKey, error) {
	if c.key_settings==nil || c.key_settings.Ext_pri==nil {
		return nil, fmt.Errorf("KeySettings invalid!")
	}

	index, err := keyToIndex(privkey, index_prefix)
	if err!=nil {
		return nil, err
	}
	return c.key_settings.Ext_pri.Child(index)
}

func (c *Client) virtualKeyToAddress(privkey string) (string, error)  {
	var address *btcutil.AddressPubKeyHash
	if extkey, err := c.virtualKeyToRealPubkey(privkey); err!=nil {
		return "", err
	}else if address, err = extkey.Address(c.chain_params);err!=nil {
		return "", err
	}
	return address.EncodeAddress(), nil
}

func (c *Client) NewAccount(co uint32) ([]*types.Account, error) {
	var index_from, index_to uint32

	c.childIndexMtx.Lock()
	defer c.childIndexMtx.Unlock()

	index_from = c.key_settings.Child_upto_index
	index_to = index_from + co

	accs := make([]*types.Account, co)

	for i:=index_from; i<index_to; i++ {
		if childpub, err := c.key_settings.Ext_pub.Child(index_from + i); err==nil {

			if config.Debugmode {
				childpri, _ := c.key_settings.Ext_pri.Child(index_from + i)

				childpri.SetNet(c.chain_params)
				depth := childpri.Depth()
				prikey, _ := childpri.ECPrivKey()
				wif, _ := btcutil.NewWIF(prikey, c.chain_params, true)
				prikey_string := wif.String()
				address, _ := childpri.Address(c.chain_params)

				l4g.Trace(`
--->Privatekey:[%d/%d], [key:%s,address:%s]`, depth, i, prikey_string, address.String() )
			}

			// converts the extended key to a standard bitcoin pay-to-pubkey-hash
			// address for the passed network.
			// AddressPubKeyHash is an Address for a pay-to-pubkey-hash (P2PKH)
			// transaction.
			if hash, err := childpub.Address(c.chain_params); err!=nil {
				l4g.Error("Convert child-extended-pub-key to address faild, message:%s", err.Error())
				return nil, err
			} else {
				if key, err := indexToKey(i, 64, index_prefix); err==nil {
					accs[i] = &types.Account{Address:hash.String(), PrivateKey:key}
				} else {
					l4g.Error("BTC Convert index to child 'private key' faild, message:%s", err.Error())
					return nil, err
				}
			}
		} else {
			l4g.Error("BTC Get child public key faild, message:%s", err.Error())
			return nil, err
		}
	}

	c.key_settings.Child_upto_index = index_to
	c.key_settings.Save()
	return accs, nil
}



