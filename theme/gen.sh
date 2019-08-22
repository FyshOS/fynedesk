#!/bin/sh

DIR=`dirname "$0"`
FILE=bundled.go
BIN=`go env GOPATH`/bin

cd $DIR
rm $FILE

$BIN/fyne bundle -package theme -name pointerDefault pointer.png > $FILE

$BIN/fyne bundle -package theme -append -name batteryIcon battery.svg >> $FILE
$BIN/fyne bundle -package theme -append -name brightnessIcon brightness.svg >> $FILE
$BIN/fyne bundle -package theme -append -name brokenImageIcon broken_image.svg >> $FILE

$BIN/fyne bundle -package theme -append -name lochFynePicture lochfyne.jpg >> $FILE
