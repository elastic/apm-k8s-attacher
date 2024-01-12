#!/bin/sh

#XXX
set -x

MAX_WAIT_SECONDS=60
POD_NAME=$1
echo "Waiting up to $MAX_WAIT_SECONDS seconds for pod $1 to start"
count=0
while [ $count -lt $MAX_WAIT_SECONDS ]
do
  count=`expr $count + 1`
  kubectl get pod -A #XXX
  STARTED=`kubectl get pod -A | grep $POD_NAME | grep 'Running'`
  if [ "$STARTED" != "" ]
  then
    exit 0
  fi
  sleep 1
done

echo "error: pod matching '$POD_NAME' failed to start within $MAX_WAIT_SECONDS seconds"
kubectl get pod -A
exit 1
