package common

import (
	"os"
	"time"
	"strings"
	"errors"
	"github.com/satori/go.uuid"
)

const(
	TimeFormat = "20060102150405"

	AddressDirName = "address"
	TxDirName = "tx"

 	MaxDbCountAddress = 10000
	MaxCountAddress = 100
)

const(
	onlineExtension = ".online"
	offlineExtension = ".offline"

	txExtension = ".tx"
	txSignedExtension = ".txs"
)

// uni address info
type UniAddressInfo struct{
	CoinType 	string
	DateTime 	string
}

// uni address db info
type UniAddressDbInfo struct{
	Uuid 		string
	OfflineMd5  string
	OnlineMd5  	string
}

func NewUniAddressInfo(coinType string) (*UniAddressInfo, error) {
	af := &UniAddressInfo{}

	af.CoinType = coinType
	af.DateTime = time.Now().Local().Format(TimeFormat)

	return af, nil
}

func NewUniAddressDbInfo() (*UniAddressDbInfo, error) {
	uadi := &UniAddressDbInfo{}

	// uuid
	uuidv4, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	uadi.Uuid = uuidv4.String()

	return uadi, nil
}

func ParseUniAddressInfo(uniDbName string) (*UniAddressInfo, error) {
	strs := strings.Split(uniDbName, "@")
	if len(strs) < 2 {
		return nil, errors.New("unidbname is error format")
	}

	af := &UniAddressInfo{}

	af.CoinType = strs[0]
	af.DateTime = strs[1]

	return af, nil
}

func (af *UniAddressInfo)MkUniAbsDir(dataDir string) error{
	return os.MkdirAll(af.GetUniAbsDir(dataDir), os.ModePerm)
}

func (af *UniAddressInfo)GetUniAbsDir(dataDir string) string{
	return dataDir + "/" + af.CoinType +"/" + af.DateTime
}

func (af *UniAddressInfo)GetUniName() string{
	return af.CoinType + "@" + af.DateTime + "@"
}

func (uadi *UniAddressDbInfo)GetUniNameOffline() string {
	return uadi.Uuid + "@" + uadi.OfflineMd5 + "@"
}

func (uadi *UniAddressDbInfo)GetUniNameOnline() string {
	return uadi.Uuid + "@" + uadi.OnlineMd5 + "@"
}

func GetOnlineExtension() string {
	return onlineExtension
}

func GetOfflineExtension() string {
	return offlineExtension
}

// uni address db info
type UniAddressLineDbInfo struct{
	CoinType 	string
	DateTime 	string
	Uuid 		string
	Md5  		string
	Ext         string
}
func ParseUniAddressLineDbInfo(uniDbName string) (*UniAddressLineDbInfo, error) {
	strs := strings.Split(uniDbName, "@")
	if len(strs) < 5 {
		return nil, errors.New("unidbname is error format")
	}

	alf := &UniAddressLineDbInfo{}

	alf.CoinType = strs[0]
	alf.DateTime = strs[1]
	alf.Uuid = strs[2]
	alf.Md5 = strs[3]
	alf.Ext = strs[4]

	return alf, nil
}

func (af *UniAddressLineDbInfo)GetUniName() string{
	return af.CoinType + "@" + af.DateTime + "@" + af.Uuid + "@" + af.Md5 + "@"
}

/////////////////////////////////////////////////////////////////////////////
// transaction format
type UniTransaction struct{
	DateTime 	string
	Uuid 		string
	Md5  		string
	Ext         string
}
func NewUniTransaction() (*UniTransaction, error) {
	uadi := &UniTransaction{}

	// uuid
	uuidv4, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	uadi.DateTime = time.Now().Local().Format(TimeFormat)
	uadi.Uuid = uuidv4.String()

	return uadi, nil
}

func ParseUniTransaction(uniDbName string) (*UniTransaction, error) {
	strs := strings.Split(uniDbName, "@")
	if len(strs) < 4 {
		return nil, errors.New("unitxname is error format")
	}

	af := &UniTransaction{}

	af.DateTime = strs[0]
	af.Uuid = strs[1]
	af.Md5 = strs[2]
	af.Ext = strs[3]

	return af, nil
}

func (af *UniTransaction)GetUniName() string{
	return af.DateTime + "@" + af.Uuid + "@" + af.Md5 + "@"
}

func GetTxExtension() string {
	return txExtension
}

func GetTxSignedExtension() string {
	return txSignedExtension
}