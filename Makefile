.PHONY: backconnectd
backconnectd: proto
	go build -o ./ ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install:
	install ./backconnectd /usr/bin/backconnectd
	install ./bcli /usr/bin/bcli

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
	rm /usr/bin/bcli
	rm -f /etc/backconnectdconfig.yml
	rm -f /etc/systemd/system/backconnectd.service

.PHONY: proto
proto:
	protoc -I=proto --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative proto/service.proto