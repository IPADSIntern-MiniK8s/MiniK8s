#!/bin/bash

# stop the components
echo "stop the minik8s"
# find the pid of the components by command
<<<<<<< HEAD
# if can find multiple pid, kill them all
pids=$(pgrep -f './apiserver')
if [[ -n "$pids" ]]; then
  echo "kill apiserver"
  for pid in $pids; do
    echo "Killing process with PID $pid"
    kill "$pid"
  done
=======
pids=$(pgrep -f './apiserver')
if [[ -n "$pids" ]]; then
  echo "kill apiserver"
  kill "$pids"
>>>>>>> remotes/origin/develop
fi

pids=$(pgrep -f './scheduler')
if [[ -n "$pids" ]]; then
  echo "kill scheduler"
  for pid in $pids; do
    echo "Killing process with PID $pid"
    kill "$pid"
  done
fi

pids=$(pgrep -f './controller')
if [[ -n "$pids" ]]; then
  echo "kill controller"
  for pid in $pids; do
    echo "Killing process with PID $pid"
    kill "$pid"
  done
fi

pids=$(pgrep -f './kubeproxy')
if [[ -n "$pids" ]]; then
  echo "kill kubeproxy"
  for pid in $pids; do
    echo "Killing process with PID $pid"
    kill "$pid"
  done
fi

pids=$(pgrep -f './serverless')
if [[ -n "$pids" ]]; then
  echo "kill serverless"
  for pid in $pids; do
    echo "Killing process with PID $pid"
    kill "$pid"
  done
fi