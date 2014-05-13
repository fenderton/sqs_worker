# heartbeat

This program acts like a wrapper around any command line program you want to run and adds a heartbeat to SQS. During the execution of the subprocess this program will update SQS every 4 minutes, using the credentials and receipt handle specified using the following, **required**, ENV variables:

* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `RECEIPT_HANDLE`
* `SQS_QUEUE_NAME`

## Usage

`./sqs_worker_heartbeat <program to run> [<arg 1> <arg 2> ...]`

## Example

`AWS_ACCESS_KEY_ID="asdf" AWS_SECRET_ACCESS_KEY="zyx" RECEIPT_HANDLE="demo" SQS_QUEUE_NAME="asdf-worker" ./sqs_worker_heartbeat /usr/bin/php /var/www/current/shell/indexer.php --info`
