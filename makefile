source:=api meta object

gen:
	$(foreach n, $(source), cd src/$(n)server; go generate ./..; cd ..)

build-all:
	$(foreach n, $(source), cd src; go build -o ../bin/$(n) $(n)server/main.go;)

start: build run

build:
	cd src
	go build -o ../bin/$(n) $(n)server/main.go

run:
	./bin/$(n) app test_conf/$(n)-server-$(i).yaml

clear:
	clear
	rm -r /workspaces/temp/*
	go test -v src/metaserver/test/api_test.go -test.run TestClearEtcd