#!/bin/sh
START_MESSAGE=`kubectl logs nodejs-test-app | grep 'agent\\\":{\\\"name\\\":\\\"nodejs\\\",'`
if [ "x$START_MESSAGE" = "x" ]
then
  exit 1
else
  exit 0
fi
