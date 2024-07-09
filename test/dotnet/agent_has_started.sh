#!/bin/sh
#initial 10 second delay for restricted CPUs making agent slow starting 
sleep 10
START_MESSAGE=`kubectl logs dotnet-test-app | grep 'Elastic APM .NET Agent, version'`
if [ "x$START_MESSAGE" = "x" ]
then
  echo "APM Agent appears to have not been started"
  exit 1
else
  echo "Found APM agent in the kubctl logs"
  exit 0
fi