all:
	go build

test:
	go test -test.run=Block
