#!/bin/bash
set -e

function print_help {
	printf "Available Commands:\n";
	awk -v sq="'" '/^function run_([a-zA-Z0-9-]*)\s*/ {print "-e " sq NR "p" sq " -e " sq NR-1 "p" sq }' make.sh \
		| while read line; do eval "sed -n $line make.sh"; done \
		| paste -d"|" - - \
		| sed -e 's/^/  /' -e 's/function run_//' -e 's/#//' -e 's/{/	/' \
		| awk -F '|' '{ print "  " $2 "\t" $1}' \
		| expand -t 30
}

function run_test { #generate embedded resources
  command -v go >/dev/null 2>&1 || { echo "executable 'go' (the language sdk) must be installed" >&2; exit 1; }
	command -v protoc >/dev/null 2>&1 || { echo "executable 'protoc' (protobuf compiler) must be installed" >&2; exit 1; }

  echo "--> generating gRPC endpoints"
	protoc --go_out=plugins=grpc:. examples/basic/*.proto

	echo "--> building gg generator"
	go build -o $GOPATH/bin/gg main.go

	echo "--> generating test examples"
	gg examples/basic/*.pb.go

	echo "--> run go tests"
	go test -v ./examples/basic
}

case $1 in
	"test") run_test ;;
	*) print_help ;;
esac
