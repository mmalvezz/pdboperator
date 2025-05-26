#!/bin/bash
export GO_ENABLED=1 
export GOOS=linux 
export GOARCH=amd64 
#export CGO_LDFLAGS=-L/usr/lib64  -Wl,--warn-once  -L/usr/lib/oracle/21/client64/lib/ -lclntsh -L./../../common
make install run WATCH_NAMESPACE="pdboperator-system,pdbnamespace" 
 



