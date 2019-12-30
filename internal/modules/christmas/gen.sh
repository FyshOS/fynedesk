#!/bin/sh

DIR=`dirname "$0"`
FILE=bundled.go
BIN=`go env GOPATH`/bin

cd $DIR
rm $FILE

$BIN/fyne bundle -package christmas lights.png > $FILE

$BIN/fyne bundle -package christmas -append tree.svg >> $FILE
