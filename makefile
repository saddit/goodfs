source:=api meta object

build-all:
	$(foreach n, $(source), go build -o bin/$(n) $(n)server/main.go;)
build-run: build run
build:
	go build -o bin/$(n) $(n)server/main.go 
run:
	./bin/$(n) app test_conf/$(n)-server-$(i).yaml
clear:
	clear
	rm -r /workspaces/temp/*