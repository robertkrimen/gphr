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
	./gphr_ release gphr_{darwin,linux,windows}_{386,amd64}*

clean:
	rm -f gphr_*
