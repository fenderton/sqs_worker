# sqs_worker

This program will get WorkOrders from a specified SQS queue and execute them in the current users context. The following ENV variables must be set:

* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `SQS_RECIEVE_QUEUE`
* `SQS_REPORT_QUEUE`

## Usage

`./sqs_worker`

## Example

`AWS_ACCESS_KEY_ID="asdf" AWS_SECRET_ACCESS_KEY="zyx" SQS_RECIEVE_QUEUE="demo" SQS_REPORT_QUEUE="asdf-worker" ./sqs_worker`
