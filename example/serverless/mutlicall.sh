#!/bin/bash

# define the serverIp
# send 3 requests to the server in a loop
# bewteen each request, sleep 0.5 seconds
# the parameter are x and y, in each loop x increase 1, y decrease 1
for i in {1..5}; do
    curl -s -X POST -H "Content-Type: application/json" -d '{"x":'$i',"y":'$((5-i))'}' http://localhost:8080/api/v1/functions/test/trigger
    sleep 0.5
done