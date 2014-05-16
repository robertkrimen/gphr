.PHONY: test build install release gphr_release clean

test: build

build:
	go build -o gphr_

install:
	go install -a

release: test
	for package in . gphr; do (cd $$package && godocdown --signature > README.markdown); done

gphr_release: test
	gnat compile .
	./gphr_ release gphr_darwin_386 gphr_linux_* gphr_windows_386.exe

clean:
	rm -f gphr_*
