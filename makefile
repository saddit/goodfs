build:
	go build $(n)server/main.go 
	mv main bin/$(n)
run:
	./bin/$(n) test_conf/$(n)-server-$(i).yaml