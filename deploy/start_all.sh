#!/bin/bash

ROOT_DIR=/home/solozyx/go/project/object-storage

services="
dbproxy
upload
download
transfer
account
apigw
"

# 替换为你的本地ip
hostIP="192.168.174.134"

registryAddr="--registry=consul --registry_address=${hostIP}:8500"
# registryAddr="--registry=kubernetes --registry_address=${hostIP}:6443"
redisAddr="--cachehost=${hostIP}:6379"
mysqlAddr="--dbhost=${hostIP}:3306"
# mqAddr="--mqhost=${hostIP}:9200"
mqAddr="--mqhost=${hostIP}:5672"

# 强制删除已有的容器
# 生产环境不建议这么做, 用k8s可以实现服务平滑重启
echo -e "\033[31m检查并停止已有的容器... \033[0m"
for service in $services
do
    app=`sudo docker ps -a | grep "hub.fileserver.com/filestore/${service}" | awk '{print $1}'`
    if [[ $app != "" ]];then
        echo $app | xargs sudo docker rm -f
    fi
done

echo -e "\033[32m启动微服务容器... \033[0m"
for service in $services
do
    volumes=""
    # 指定挂载目录
    if [[ $service == "upload" || $service == "download" ]];then
        volumes="-v /data/fileserver:/data/fileserver -v /data/fileserver_part:/data/fileserver_part"
    fi

    # 启动容器
    sudo docker run -it -d \
      --net=host --privileged=true ${volumes} \
      -e PARAMS="${registryAddr} ${redisAddr} ${mysqlAddr} ${mqAddr}" \
      hub.fileserver.com/filestore/${service}
done
