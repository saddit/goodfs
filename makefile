source := api meta object admin
i := 1

define build-ui
	cd src/adminserver/ui; yarn run build
endef

gen:
	$(foreach n, $(source), go generate src/$(n)server/..;)
	go generate src/common/..

build-all:
	$(build-ui)
	$(foreach n, $(source), go build -o bin/$(n) src/$(n)server/main.go;)

start: build run

yarn-build:
	$(build-ui)

yarn-dev:
	cd src/adminserver/ui; yarn run dev

yarn-init:
	cd src/adminserver/ui; yarn

build:
ifeq ($(n),admin)
	$(build-ui)
endif
	go build -o bin/$(n) src/$(n)server/main.go

run:
	./bin/$(n) app test_conf/$(n)-server-$(i).yaml

clear:
	clear
	rm -rf /workspaces/temp/*
	go test -v src/metaserver/test/api_test.go -test.run TestClearEtcd