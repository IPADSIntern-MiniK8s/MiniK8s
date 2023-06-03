#!/bin/bash

# cd to the home directory
current_path=$(pwd)
cd ~
cd /home

# check if the ectd is running, if not, start it in the background
# etcd is a progress
if ! pgrep -x "etcd" > /dev/null
then
    echo "etcd is not running, start it"
    nohup etcd &
fi

# check the default systemd-resolved, if it is running, stop it
if pgrep -x "systemd-resolved" > /dev/null
then
    echo "systemd-resolved is running, stop it"
    systemctl stop systemd-resolved
fi

# check if the coredns is running, if not, start it in the background
if ! pgrep -x "coredns" > /dev/null
then
    echo "coredns is not running, start it"
    nohup ./coredns -conf $(pwd)/mini-k8s/pkg/kubedns/config/Corefile &
fi

# check the default nginx, if it is running, stop it
if pgrep -x "nginx" > /dev/null
then
    echo "nginx is running, stop it"
    systemctl stop nginx
fi

# start the nginx in the background
echo "start nginx"
nohup nginx -c $(pwd)/mini-k8s/pkg/kubedns/config/nginx.conf &

# build the components and run the server
cd "$current_path"
make kubectl
make apiserver
make scheduler
make controller
make serverless
make kubeproxy

# create the log directory if not exist
if [ ! -d "./log" ]; then
  mkdir ./log
fi


cd bin

# start the components in different terminals
echo "start the minik8s"
# ./apiserver > ../log/apiserver.log 2> /dev/null &

./apiserver > ../log/apiserver.log 2>&1 &
echo "start apiserver"
sleep 3
./scheduler > ../log/scheduler.log 2>&1 &
echo "start scheduler"
./controller > ../log/controller.log 2>&1 &
echo "start controller"
./kubeproxy > ../log/kubeproxy.log 2>&1 &


chmod +x ../../pkg/serverless/function/registry.sh
cd ../../pkg/serverless/function
chmod +x ../../
./registry.sh
cd ../../../build/bin/
./serverless  > ../log/serverless.log 2>&1 &
echo "start serverless"


