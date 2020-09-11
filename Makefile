all:
	go build -gcflags "-N -l" -o search search.go
