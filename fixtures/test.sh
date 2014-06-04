#!/bin/bash

echo "starting..."

echo "got arg0: $0"
echo "got arg1: $1"
echo "got arg2: $2"
echo "got arg3: $3"

echo `cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 300000 | head -n 1`

echo "CWD: `pwd`"

sleep $2

echo "done"

echo `date`

exit 0
