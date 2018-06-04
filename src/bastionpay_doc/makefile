all: clean media

clean:
	@rm -rf ./bastionpay_doc*

media:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./bastionpay_doc