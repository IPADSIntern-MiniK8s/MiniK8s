#!/bin/bash

# define the serverIp
# send 3 requests to the server in a loop
# bewteen each request, sleep 0.5 seconds
# the parameter are x and y, in each loop x increase 1, y decrease 1
cd ../../build/bin/
for i in {1..4}; do
    ./kubectl trigger function test -f /home/mini-k8s/example/serverless/param.yaml >> ../log/output.log &
done