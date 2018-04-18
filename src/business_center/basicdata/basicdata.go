package basicdata

import (
	. "business_center/def"
	"business_center/mysqlpool"
	"sync"
)

var ins *BasicData = nil

func init() {
	b := new(BasicData)

	b.mapUserProperty = make(map[string]*UserProperty)
	b.mapAssetProperty = make(map[string]*AssetProperty)
	b.mapUserAddress = make(map[string]*UserAddress)

	b.mapUserProperty, _ = mysqlpool.QueryAllUserProperty()
	b.mapAssetProperty, _ = mysqlpool.QueryAllAssetProperty()
	b.mapUserAddress, _ = mysqlpool.QueryAllUserAddress()

	ins = b
}

func Get() *BasicData {
	return ins
}

type BasicData struct {
	mapUserProperty  map[string]*UserProperty
	mapAssetProperty map[string]*AssetProperty
	mapUserAddress   map[string]*UserAddress

	lockUserPropertyMap  sync.RWMutex
	lockAssetPropertyMap sync.RWMutex
	lockUserAddressMap   sync.RWMutex
}

func (b *BasicData) GetAllUserPropertyMap() map[string]*UserProperty {
	b.lockUserPropertyMap.RLock()
	defer b.lockUserPropertyMap.RUnlock()
	return b.mapUserProperty
}

func (b *BasicData) AddAssetPropertyMap(data []AssetProperty) {
	b.lockUserPropertyMap.Lock()
	defer b.lockUserPropertyMap.Unlock()

	for _, v := range data {
		b.mapAssetProperty[v.Name] = &v
	}
}

func (b *BasicData) GetAllAssetPropertyMap() map[string]*AssetProperty {
	b.lockAssetPropertyMap.RLock()
	defer b.lockAssetPropertyMap.RUnlock()
	return b.mapAssetProperty
}

func (b *BasicData) AddUserPropertyMap(data []UserProperty) {
	b.lockAssetPropertyMap.Lock()
	defer b.lockAssetPropertyMap.Unlock()

	for _, v := range data {
		b.mapUserProperty[v.UserKey] = &v
	}
}

func (b *BasicData) GetAllUserAddressMap() map[string]*UserAddress {
	b.lockUserAddressMap.RLock()
	defer b.lockUserAddressMap.RUnlock()
	return b.mapUserAddress
}

func (b *BasicData) AddUserAddressMap(data []UserAddress) {
	b.lockUserAddressMap.Lock()
	defer b.lockUserAddressMap.Unlock()

	for _, v := range data {
		b.mapUserAddress[v.AssetName+"_"+v.Address] = &v
	}
}
