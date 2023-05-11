#!/bin/sh
START_MESSAGE=`kubectl logs java-test-app | grep 'agent.configuration.StartupInfo - Starting Elastic APM'`
if [ "x$START_MESSAGE" = "x" ]
then
  exit 1
else
  exit 0
fi
