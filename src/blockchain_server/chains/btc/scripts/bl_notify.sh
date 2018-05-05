#!/bin/sh
curl -d "blhash=$1" http://127.0.0.1:18666/btc/bl_notify
