deps:
	go get github.com/spacemonkeygo/openssl
build:
	make clean
	mkdir ./bin
	GOOS=linux go build -o ./bin/groxy-server ./server/server.go
	GOOS=linux go build -o ./bin/groxy-client ./client/client.go
clean:
	rm -rf ./bin
	rm -f ./client.log
	rm -f ./server.log