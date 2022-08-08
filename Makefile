.PHONY: build
build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -v -o bin/tmpl8 . 

test1:
	cat test/input.json | go run . -v test/templates/* | tee test/input1.result

test2:
	find test/templates -type f > test/files.lst
	cat test/input.json | go run . @test/files.lst | tee test/input2.result

test3: build
	find test/templates -type f > test/files.lst
	echo k8s:default/tmpl8-demo >> test/files.lst
	cat test/input.json | ./tmpl8 -v @test/files.lst | tee test/input2.result

install: build
	cp bin/tmpl8 ~/bin

k8:
	kubectl apply -k test/k8s

m-y:
	go run . <test/m-input.yaml test/m-tmpl.yaml -v

m-j:
	go run . -s <test/m-input.json test/m-tmpl.yaml -v

multi:
	go run . -i test/m-input.yaml -i test/m-input.json test/m-tmpl.yaml -v -s

adv:
	go run . -i test/advanced/input.yaml test/advanced/*.tmpl8 -v

check:
	@echo "Checking...\n"
	gocyclo -over 15 . || echo -n ""
	@echo ""
	golangci-lint run -E misspell -E depguard -E dupl -E goconst -E gocyclo -E ifshort -E predeclared -E tagliatelle -E errorlint -E godox -D structcheck
	@echo ""
	golint -min_confidence 0.21 -set_exit_status ./...
	@echo "\nAll ok!"

release:
	gh release create $(TAG) -t $(TAG)