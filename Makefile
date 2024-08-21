.PHONY: fonts-clean

all: fonts/Aileron-Regular.otf
	rm -rf bin
	GOOS=darwin GOARCH=arm64 go build -o bin/darwin-arm64/imageclock
	GOOS=linux GOARCH=arm64 go build -o bin/linux-arm64/imageclock
	GOOS=linux GOARCH=amd64 go build -o bin/linux-amd64/imageclock

fonts/Aileron-Regular.otf:
	mkdir -p ./fonts
	wget https://www.fontsquirrel.com/fonts/download/aileron -O fonts/aileron
	unzip fonts/aileron -d fonts
	rm fonts/aileron


fonts-clean:
	rm -rf fonts
