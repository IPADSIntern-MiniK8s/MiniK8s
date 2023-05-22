#!/bin/bash

# define the serverIp
serverIp="192.168.1.13"

# if not have registry, then pull from docker hub
if ! docker images | grep -q "registry"; then
  docker pull registry
fi

# if the registry container is not running, then start it
if ! docker ps | grep -q "registry"; then
  docker run -d -p 5000:5000 --restart=always --name registry registry
fi

# if the current ip not equal to the server ip, then mark the server ip as trusted
if [ "$(hostname -I | awk '{print $1}')" != "$serverIp" ]; then
  # check whether the file exist
  if [ ! -f /etc/docker/daemon.json ]; then
      touch /etc/docker/daemon.json
  fi

  # add http 
  echo '{
    "insecure-registries": ["'"$serverIp:5000"'"]
  }' > /etc/docker/daemon.json

  systemctl restart docker
fi


