package btc

import (
	"encoding/binary"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcwallet/wtxmgr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"time"
	"encoding/hex"
	"strconv"
	"blockchain_server/utils"
)

func indexToKey (index uint32, tolen uint, slat string) (string, error) {
	if tolen < 64 {tolen=64}

	index_hex := strconv.FormatUint(uint64(index), 16)

	md5 := utils.MD5(slat + strconv.FormatInt(int64(index), 16))

	index_len := len(index_hex)
	bs := make([]byte, 4)
	// 用32位, 4个字节, 转换成16进制, 形成字符串, 表示index的字符串的位数
	// 字符串需要站8个字符的位置!!
	binary.LittleEndian.PutUint32(bs, uint32(index_len))

	rdlen := tolen - (uint(len(index_hex + md5)) + 8)
	rdstr := utils.RandString(int(rdlen))

	return md5 + fmt.Sprintf("%08x", bs) + rdstr + index_hex, nil
}
// parseBlock parses a btcws definition of the block a tx is mined it to the
// Block structure of the wtxmgr package, and the block index.  This is done
// here since rpcclient doesn't parse this nicely for us.
func parseBlock(block *btcjson.BlockDetails) (*wtxmgr.BlockMeta, error) {
	if block == nil {
		return nil, nil
	}
	blkHash, err := chainhash.NewHashFromStr(block.Hash)
	if err != nil {
		return nil, err
	}
	blk := &wtxmgr.BlockMeta{
		Block: wtxmgr.Block{
			Height: block.Height,
			Hash:   *blkHash,
		},
		Time: time.Unix(block.Time, 0),
	}
	return blk, nil
}

func keyToIndex (index_str, slat string) (uint32, error) {
	// TODO: to check if data has been change by some one!
	index_len_hex := index_str[32:40]

	if index_len_bs, err := hex.DecodeString(index_len_hex); err!=nil {
		return 0, err
	} else {
		index_len := binary.LittleEndian.Uint32(index_len_bs)
		real_index_str := string(index_str[64-index_len:])
		if index, err := strconv.ParseUint(real_index_str, 16, 32); err!=nil {
			return 0, err
		} else { return uint32(index), nil }
	}
}

