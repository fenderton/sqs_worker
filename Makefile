.PHONY: build

build:
	docker run --rm -v $(PWD):/usr/src/sqs_worker -w /usr/src/sqs_worker golang:1.4 make docker-compile

docker-compile:
	go get github.com/crowdmob/goamz/sqs
	go get github.com/ianneub/logger
	go build -v

test:
	docker run --rm -v $(PWD):/go/src/sqs_worker -w /go/src/sqs_worker golang:1.4 make docker-test

docker-test:
	go get github.com/crowdmob/goamz/sqs
	go get github.com/ianneub/logger
	go test

run:
	docker run --rm -it -v $(PWD):/usr/src/sqs_worker --env-file=.env centos /usr/src/sqs_worker/sqs_worker -d
