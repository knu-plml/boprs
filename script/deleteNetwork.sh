#!/bin/sh

docker rm -f $(docker ps -qa)
#docker image rm $(docker images --filter=reference="dev-*" -q)
