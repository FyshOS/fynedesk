#!/bin/bash
PREFIX=/usr/local

cd "$(dirname "$0")"

# build
cd cmd/fynedesk_runner
go build .
if [ $? -ne 0 ]; then
	echo "Failed to build fynedesk_runner"
	exit 1
fi

cd ../fynedesk
go build .
if [ $? -ne 0 ]; then
	echo "Failed to build fynedesk"
	exit 1
fi

cd ../..

#install
echo "Installing to $PREFIX/bin and /usr/local/xsessions..."
sudo -- sh -c "cp cmd/fynedesk_runner/fynedesk_runner $PREFIX/bin/ && \
	cp cmd/fynedesk/fynedesk $PREFIX/bin/ && \
	cp fynedesk.desktop /usr/share/xsessions"

#done
echo "Install completed successfully"
