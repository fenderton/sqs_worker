FROM ianneub/go
RUN go get github.com/crowdmob/goamz/sqs
RUN go get github.com/ianneub/logger
