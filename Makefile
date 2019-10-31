EXT=
ifeq (${GOOS},windows)
    EXT=.exe
endif

.PHONY: all
all: clean build install

.PHONY: clean build install

build:
	go build -o ./build/gclient .

install: build
	install ./build/gclient ${GOPATH}/bin/gclient${EXT}	

clean:
	rm -rf ./build 