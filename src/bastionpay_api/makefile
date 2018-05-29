all: clean media

clean:
	@rm -rf ./bastionpay_api_darwin*
	@rm -rf ./bastionpay_api_linux*
	@rm -rf ./bastionpay_api_windows.exe*

media:
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -v -o ./bastionpay_api_darwin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bastionpay_api_linux
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -v -o ./bastionpay_api_windows.exe
