install:
	go build -ldflags '-s -w' -o $(CURDIR)/build/bin/cfy .
	chmod +x $(CURDIR)/build/bin/cfy
	cp $(CURDIR)/build/bin/cfy /usr/local/bin
	rm $(CURDIR)/build/bin/cfy