#!/bin/sh

name=testNode              # Default node name.
net=testNetwork            # Default docker network name.
image=default-fabric-image # Default ubuntu image name.

help() {
  echo "createCouch.sh [OPTIONS]"
  echo "               -h           Explain options."
  echo "               -c <string>  Setting docker container name."
  echo "               -n <string>  Setting docker network name."
  echo "               -i <string>  Create container using docker image name."
  exit 0
}

while getopts "c:n:i:h" opt
do
  case $opt in
    c) name=$OPTARG;;
    n) net=$OPTARG;;
    i) image=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

docker run -itd --name $name --net=$net -e COUCHDB_USER="" -e COUCHDB_PASSWORD="" $image /bin/bash
echo "A node which name is ${name} in ${net} using ${image} is created."
