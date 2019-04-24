#!/bin/sh

DIR=`dirname "$0"`
FILE=bundled.go
BIN=`go env GOPATH`/bin

cd $DIR
rm $FILE

$BIN/fyne bundle -package theme -name pointerDefault pointer.png > $FILE

$BIN/fyne bundle -package theme -append -name batteryLight battery_light.svg >> $FILE
$BIN/fyne bundle -package theme -append -name batteryDark battery_dark.svg >> $FILE
$BIN/fyne bundle -package theme -append -name brightnessLight brightness_light.svg >> $FILE
$BIN/fyne bundle -package theme -append -name brightnessDark brightness_dark.svg >> $FILE
