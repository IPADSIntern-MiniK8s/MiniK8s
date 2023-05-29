#!/bin/bash

# cd to the home directory
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
    nohup ./coredns -conf /home/mini-k8s/pkg/kubedns/config/Corefile &
fi

# check the default nginx, if it is running, stop it
if pgrep -x "nginx" > /dev/null
then
    echo "nginx is running, stop it"
    systemctl stop nginx
fi

# start the nginx in the background
echo "start nginx"
nohup nginx -c /home/mini-k8s/pkg/kubedns/config/nginx.conf &

# build the components and run the server
cd /home/mini-k8s/build
make apiserver
make scheduler
make controller
make serverless
make kubeproxy
cd bin

# start the components in different terminals
echo "start the minik8s"
# ./apiserver > ../log/apiserver.log 2> /dev/null &
./apiserver > ../log/apiserver.log 2>&1 &
sleep 3
./scheduler > ../log/scheduler.log 2>&1 &
./controller > ../log/controller.log 2>&1 &
./kubeproxy > ../log/kubeproxy.log 2>&1 &


chmod +x /home/mini-k8s/pkg/serverless/function/registry.sh
cd /home/mini-k8s/pkg/serverless/function
./registry.sh
cd /home/mini-k8s/build/bin
./serverless  > ../log/serverless.log 2>&1 &



