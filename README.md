# sqs_worker

This program will get WorkOrders from a specified SQS queue and execute them in the current users context. The following ENV variables must be set:

* `SQS_WORKER_ACCESS_KEY`
* `SQS_WORKER_SECRET_KEY`
* `SQS_WORKER_QUEUE`
* `SQS_REPORT_QUEUE`
* `CMD_BASE`
* `CMD_DIR`

## Usage

`./sqs_worker`

## Example

`SQS_WORKER_ACCESS_KEY="asdf" SQS_WORKER_SECRET_KEY="zyx" SQS_WORKER_QUEUE="demo" SQS_REPORT_QUEUE="asdf-worker" CMD_BASE="/usr/bin/php" CMD_DIR="/var/www/current/shell" ./sqs_worker`
