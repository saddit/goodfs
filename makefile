source:=api meta object admin

gen:
	$(foreach n, $(source), cd src/$(n)server; go generate ./..; cd ..)

build-all: build-yarn
	$(foreach n, $(source), go build -o bin/$(n) src/$(n)server/main.go;)

build-yarn:
	cd src/adminserver/ui; yarn

start: build run

build:
	ifeq ($(n),'admin')
		build-yarn
	go build -o bin/$(n) src/$(n)server/main.go

run:
	./bin/$(n) app test_conf/$(n)-server-$(i).yaml

clear:
	clear
	rm -rf /workspaces/temp/*
	go test -v src/metaserver/test/api_test.go -test.run TestClearEtcd