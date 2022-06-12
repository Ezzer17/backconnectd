.PHONY: backconnectd
backconnectd:
	go build -o ./ ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install:
	install ./backconnectd /usr/bin/backconnectd

.PHONY: install-service
install-service:
	install ./config.yml /etc/backconnectdconfig.yml
	install ./systemd/backconnectd.service /etc/systemd/system/backconnectd.service

.PHONY: clean
clean:
	rm ./backconnectd

.PHONY: uninstall
uninstall:
	rm /usr/bin/backconnectd
	rm -f /etc/backconnectdconfig.yml
	rm -f /etc/systemd/system/backconnectd.service