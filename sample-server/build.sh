rm -rf test_client test_server
go build -o bin/test_client src/client/cmd/client.go
go build -o bin/test_server src/server/cmd/server.go
chmod 777 bin/test_client
chmod 777 bin/test_server