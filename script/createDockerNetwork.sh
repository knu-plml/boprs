#!/bin/sh

net=testNetwork # Default docker network name.

help() {
  echo "createDockerNetwork.sh [OPTIONS]"
  echo "                       -h           Explain options."
  echo "                       -n <string>  Setting docker network name."
  exit 0
}

while getopts "n:h" opt
do
  case $opt in
    n) net=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

docker network create --driver=bridge $net
echo "A docker network which name is ${net} is created."
