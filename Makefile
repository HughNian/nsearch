BINPATH=./
BINNAME=nsearch

ifeq ($(OS),Windows_NT)
	PLATFORM=windows
	BINNAME=nsearch.exe
else
	ifeq ($(shell uname),Darwin)
		PLATFORM=darwin
	else
		PLATFORM=linux
	endif
endif

all:
	GO111MODULE=on GOOS=$(PLATFORM) CGO_ENABLED=0 go build -gcflags "-N -l" -o $(BINPATH)$(BINNAME) search.go
