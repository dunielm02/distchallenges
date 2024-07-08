echo:
	set GOARCH=amd64
	set GOOS=linux
	go build -o ./maelstrom/maelstrom-echo ./cmd/echo/main.go

echo-test:
	./maelstrom/maelstrom test -w echo --bin ./maelstrom/maelstrom-echo --node-count 1 --time-limit 10

generate-id:
	set GOARCH=amd64
	set GOOS=linux
	go build -o ./maelstrom/maelstrom-unique-ids ./cmd/unique-id/main.go

generate-id-test:
	./maelstrom/maelstrom test -w unique-ids --bin ./maelstrom/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

broadcast:
	set GOARCH=amd64
	set GOOS=linux
	go build -o ./maelstrom/maelstrom-broadcast ./cmd/broadcast/main.go

broadcast-test:
	./maelstrom/maelstrom test -w broadcast --bin ./maelstrom/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10