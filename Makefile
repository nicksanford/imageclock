.PHONY: fonts-clean

all: clockdrawer/fonts/Aileron-Regular.otf
	rm -rf bin
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/imageclock
	GOOS=linux GOARCH=arm64 go build -o bin/linux-arm64/imageclock
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/imageclock

clockdrawer/fonts/Aileron-Regular.otf:
	mkdir -p ./clockdrawer/fonts
	wget https://www.fontsquirrel.com/fonts/download/aileron -O clockdrawer/fonts/aileron
	unzip clockdrawer/fonts/aileron -d clockdrawer/fonts
	rm clockdrawer/fonts/aileron


fonts-clean:
	rm -rf clockdrawer/fonts
