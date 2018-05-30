package chain

import "blockchain_server/service"

var wallet *service.ClientManager = nil

func SetWallet(w *service.ClientManager) {
	wallet = w
}
