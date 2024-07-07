echo:
	set GOARCH=amd64
	set GOOS=linux
	go build -o ./maelstrom/maelstrom-echo ./cmd/echo/main.go

echo-test:
	./maelstrom/maelstrom test -w echo --bin ./maelstrom/maelstrom-echo --node-count 1 --time-limit 10
