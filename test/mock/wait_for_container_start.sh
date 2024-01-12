#!/bin/sh

MAX_WAIT_SECONDS=60
POD_NAME=$1

echo "Waiting up to $MAX_WAIT_SECONDS seconds for pod $POD_NAME to start"
count=0
while [ $count -lt $MAX_WAIT_SECONDS ]
do
  count=`expr $count + 1`
  STARTED=`kubectl get pod -A | grep $POD_NAME | grep 'Running'`
  if [ "$STARTED" != "" ]
  then
    exit 0
  fi
  sleep 1
done

echo "error: pod matching '$POD_NAME' failed to start within $MAX_WAIT_SECONDS seconds"
echo "-- pod info:"
kubectl get pod -A
echo "--"
exit 1
