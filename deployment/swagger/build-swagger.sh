#!/usr/bin/env bash

#set -ex

docker build -t swagfromhome .

old_id=$(docker ps | grep :3200 | awk {'print$1'})
echo "${old_id}"

if [  -n "$old_id" ]; then
docker rm -f "$old_id"
fi

docker rmi $(docker images | grep none | awk '{print$3}')


docker run -d -p 8080:8080 swagfromhome

echo "-------------------------------"
echo "New container is running:"
docker ps | grep swagfromhome