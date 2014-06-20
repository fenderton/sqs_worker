FROM ianneub/go:1.3

RUN go get github.com/crowdmob/goamz/sqs
RUN go get github.com/ianneub/logger
