#!/bin/bash

# stop the components
echo "stop the minik8s"
# find the pid of the components by command
pids=$(pgrep -f './apiserver')
if [[ -n "$pids" ]]; then
  echo "kill apiserver"
  kill "$pids"
fi

pids=$(pgrep -f './scheduler')
if [[ -n "$pids" ]]; then
  echo "kill scheduler"
  kill "$pids"
fi

pids=$(pgrep -f './controller')
if [[ -n "$pids" ]]; then
  echo "kill controller"
  kill "$pids"
fi

pids=$(pgrep -f './serverless')
if [[ -n "$pids" ]]; then
  echo "kill serverless"
  kill "$pids"
fi