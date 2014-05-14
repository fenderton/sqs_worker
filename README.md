# sqs_worker

This program will get WorkOrders from a specified SQS queue and execute them in the current users context. The following ENV variables must be set:

* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `SQS_RECIEVE_QUEUE`
* `SQS_REPORT_QUEUE`
* `CMD_BASE`
* `CMD_DIR`

## Usage

`./sqs_worker`

## Example

`AWS_ACCESS_KEY_ID="asdf" AWS_SECRET_ACCESS_KEY="zyx" SQS_RECIEVE_QUEUE="demo" SQS_REPORT_QUEUE="asdf-worker" CMD_BASE="/usr/bin/php" CMD_DIR="/var/www/current/shell" ./sqs_worker`
