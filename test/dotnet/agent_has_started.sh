#!/bin/sh
#initial 10 second delay for restricted CPUs making agent slow starting 
sleep 10
START_MESSAGE=`kubectl logs dotnet-test-app | grep '- Elastic APM .NET Agent, version:'`
if [ "x$START_MESSAGE" = "x" ]
then
  exit 1
else
  exit 0
fi