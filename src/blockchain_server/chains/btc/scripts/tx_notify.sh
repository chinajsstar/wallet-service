#!/bin/sh
curl -d "txid=$1" http://127.0.0.1:18666/btc/tx_notify
