#!/bin/ash

# write out the env
env
# make it so we can stop the container
trap exit SIGINT SIGQUIT SIGTERM
# and wait
sleep 1d
