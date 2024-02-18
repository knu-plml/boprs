#!/bin/sh

name=testNode              # Default node name.
net=testNetwork            # Default docker network name.
image=default-fabric-image # Default ubuntu image name.
port=0                     # Default port

help() {
  echo "createNode.sh [OPTIONS]"
  echo "              -h               Explain options."
  echo "              -c <string>      Setting docker container name."
  echo "              -n <string>      Setting docker network name."
  echo "              -i <string>      Create container using docker image name."
  echo "              -p <Cport:Hport> Setting Container port: Host Port."
  exit 0
}

while getopts "c:n:i:p:h" opt
do
  case $opt in
    c) name=$OPTARG;;
    n) net=$OPTARG;;
    i) image=$OPTARG;;
    p) port=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

if [ $port = 0 ]; then
docker run -itd --name $name --net=$net -e GOPATH="/root/gopath" -e GOROOT="/root/go" -e FABRIC_HOME="/root/gopath/src/github.com/hyperledger/fabric" -e PATH="$PATH:/root/gopath/bin:/root/go/bin:/root/gopath/src/github.com/hyperledger/fabric/build/bin:/root/gopath/src/github.com/hyperledger/fabric-ca/bin" -e FABRIC_CA_CLIENT_HOME="/root/testnet" -e FABRIC_CA_SERVER_HOME="/root/testnet" -e FABRIC_CFG_PATH="/root/testnet" $image /bin/bash
else
docker run -itd --name $name --net=$net -e GOPATH="/root/gopath" -e GOROOT="/root/go" -e FABRIC_HOME="/root/gopath/src/github.com/hyperledger/fabric" -e PATH="$PATH:/root/gopath/bin:/root/go/bin:/root/gopath/src/github.com/hyperledger/fabric/build/bin:/root/gopath/src/github.com/hyperledger/fabric-ca/bin" -e FABRIC_CA_CLIENT_HOME="/root/testnet" -e FABRIC_CA_SERVER_HOME="/root/testnet" -e FABRIC_CFG_PATH="/root/testnet" -p $port/tcp $image /bin/bash
fi
docker exec -itd $name mkdir /root/testnet
if [ -f ./tls-cert.pem ]; then
  docker cp ./tls-cert.pem $name:/root/testnet
else
  echo "tls-cert.pem does not exist"
fi
echo "A node which name is ${name} in ${net} using ${image} is created."
