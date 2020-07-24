#!/bin/sh

DIR=`dirname "$0"`
FILE=bundled.go
BIN=`go env GOPATH`/bin

cd $DIR
rm $FILE

$BIN/fyne bundle -package theme -name pointerDefault pointer.png > $FILE

$BIN/fyne bundle -package theme -append -name batteryIcon battery.svg >> $FILE
$BIN/fyne bundle -package theme -append -name brightnessIcon brightness.svg >> $FILE
$BIN/fyne bundle -package theme -append -name calculateIcon calculate.svg >> $FILE
$BIN/fyne bundle -package theme -append -name displayIcon display.svg >> $FILE
$BIN/fyne bundle -package theme -append -name personIcon person.svg >> $FILE
$BIN/fyne bundle -package theme -append -name internetIcon internet.svg >> $FILE


$BIN/fyne bundle -package theme -append -name brokenImageIcon broken_image.svg >> $FILE
$BIN/fyne bundle -package theme -append -name maximizeIcon maximize.svg >> $FILE
$BIN/fyne bundle -package theme -append -name iconifyIcon minimize.svg >> $FILE
$BIN/fyne bundle -package theme -append -name keyboardIcon keyboard.svg >> $FILE
$BIN/fyne bundle -package theme -append -name soundIcon sound.svg >> $FILE
$BIN/fyne bundle -package theme -append -name muteIcon mute.svg >> $FILE

$BIN/fyne bundle -package theme -append -name lochFynePicture lochfyne.jpg >> $FILE
$BIN/fyne bundle -package theme -append -name fyneAboutBackground fyne_about_bg.png >> $FILE
