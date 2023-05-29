# use basic image
# refer from zhihu, it is the fastest way to build image
# FROM ubuntu:20.04
# RUN apt-get update && apt-get install -y python3 python3-pip
FROM python:3.9-slim

COPY . .
# set workdir
WORKDIR /app

EXPOSE 8081

# copy file into workdir
copy . .
# install dependencies from requirements.txt
RUN pip3 install -i https://pypi.tuna.tsinghua.edu.cn/simple --default-timeout=60 --no-cache-dir -r ./requirements.txt

# set env
ENV PATH="/usr/local/bin:${PATH}"

# run server.py, start the function server
CMD ["python3", "server.py"]