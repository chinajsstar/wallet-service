#!/bin/sh
curl -d "alert=$1" http://127.0.0.1:18666/btc/al_notify
