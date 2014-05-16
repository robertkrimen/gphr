.PHONY: test build install release clean

test: build

build:
	go build -o gphr_

install:
	go install -a

release: test
	for package in . gphr; do (cd $$package && godocdown --signature > README.markdown); done

clean:
	rm -f gphr_
