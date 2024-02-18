#!/bin/sh

name=testNode              # Default node name.
net=testNetwork            # Default docker network name.
image=default-fabric-image # Default ubuntu image name.
port=0                     # Default port
eport=0                    # Default event port

help() {
  echo "createNodeDooD.sh [OPTIONS]"
  echo "                  -h               Explain options."
  echo "                  -c <string>      Setting docker container name,
                                   This docker container use Docker Out of Docker(DooD)
                                   using v option in docker command because
                                   container needs to execute docker in docker container
                                   for instantiating fabric chaincode."
  echo "                  -n <string>      Setting docker network name."
  echo "                  -i <string>      Create container using docker image name."
  echo "                  -p <Cport:Hport> Setting Container port: Host Port."
  echo "                  -e <Cport:Hport> Setting Container port: Host Port."
  exit 0
}

while getopts "c:n:i:p:e:h" opt
do
  case $opt in
    c) name=$OPTARG;;
    n) net=$OPTARG;;
    i) image=$OPTARG;;
    p) port=$OPTARG;;
    e) eport=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

if [ $port = 0 ]; then
docker run -itd -v /var/run/docker.sock:/var/run/docker.sock --name $name --net=$net -e CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE="$net" -e CORE_VM_ENDPOINT="unix:///var/run/docker.sock" -e GOPATH="/root/gopath" -e GOROOT="/root/go" -e FABRIC_HOME="/root/gopath/src/github.com/hyperledger/fabric" -e PATH="$PATH:/root/gopath/bin:root/go/bin:/root/gopath/src/github.com/hyperledger/fabric/build/bin:/root/gopath/src/github.com/hyperledger/fabric-ca/bin" -e FABRIC_CA_CLIENT_HOME="/root/testnet" -e FABRIC_CA_SERVER_HOME="/root/testnet" -e FABRIC_CFG_PATH="/root/testnet" $image /bin/bash
elif [ $eport = 0 ]; then
docker run -itd -v /var/run/docker.sock:/var/run/docker.sock --name $name --net=$net -e CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE="$net" -e CORE_VM_ENDPOINT="unix:///var/run/docker.sock" -e GOPATH="/root/gopath" -e GOROOT="/root/go" -e FABRIC_HOME="/root/gopath/src/github.com/hyperledger/fabric" -e PATH="$PATH:/root/gopath/bin:root/go/bin:/root/gopath/src/github.com/hyperledger/fabric/build/bin:/root/gopath/src/github.com/hyperledger/fabric-ca/bin" -e FABRIC_CA_CLIENT_HOME="/root/testnet" -e FABRIC_CA_SERVER_HOME="/root/testnet" -e FABRIC_CFG_PATH="/root/testnet" -p $port/tcp $image /bin/bash
else
docker run -itd -v /var/run/docker.sock:/var/run/docker.sock --name $name --net=$net -e CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE="$net" -e CORE_VM_ENDPOINT="unix:///var/run/docker.sock" -e GOPATH="/root/gopath" -e GOROOT="/root/go" -e FABRIC_HOME="/root/gopath/src/github.com/hyperledger/fabric" -e PATH="$PATH:/root/gopath/bin:/root/go/bin:/root/gopath/src/github.com/hyperledger/fabric/build/bin:/root/gopath/src/github.com/hyperledger/fabric-ca/bin" -e FABRIC_CA_CLIENT_HOME="/root/testnet" -e FABRIC_CA_SERVER_HOME="/root/testnet" -e FABRIC_CFG_PATH="/root/testnet" -p $port/tcp -p $eport/tcp $image /bin/bash
fi
docker exec -itd $name mkdir /root/testnet
if [ -f ./tls-cert.pem ]; then
  docker cp ./tls-cert.pem $name:/root/testnet
else
  echo "tls-cert.pem does not exist"
fi
echo "A node using DooD which name is ${name} in ${net} using ${image} is created."
