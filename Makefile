.PHONY: test release install build clean

test:
	$(MAKE) -C gphr $@

release: test
	for package in . gphr; do (cd $$package && godocdown --signature > README.markdown); done

install: test
	$(MAKE) -C gphr $@
	go install

build:
	$(MAKE) -C gphr $@

clean:
	$(MAKE) -C gphr $@
