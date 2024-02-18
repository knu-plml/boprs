#!/bin/sh

docker build -t init_ubuntu .

docker run -itd -v /var/run/docker.sock:/var/run/docker.sock --name init_ubuntu init_ubuntu

docker exec -w /root/gopath/src/github.com/hyperledger/fabric -it init_ubuntu make

docker exec -w /root/gopath/ -it init_ubuntu bash -c "go get -u github.com/hyperledger/fabric-ca/cmd/..."
docker exec -w /root/gopath/ -it init_ubuntu bash -c "go get github.com/hyperledger/fabric-chaincode-go/shim"
#docker exec -w /root/gopath/ -it init_ubuntu bash -c "go get github.com/hyperledger/fabric-ca/cmd/..."

#docker exec -w /root/gopath/src/github.com/hyperledger/fabric-ca -it init_ubuntu make

docker exec -it init_ubuntu bash -c "cp /root/gopath/src/github.com/hyperledger/fabric/sampleconfig/core.yaml /root/testnet/"
docker exec -it init_ubuntu bash -c "cp /root/gopath/src/github.com/hyperledger/fabric/sampleconfig/orderer.yaml /root/testnet/"
docker exec -it init_ubuntu bash -c "cp /root/gopath/src/github.com/hyperledger/fabric/sampleconfig/configtx.yaml /root/testnet/"

docker commit -p init_ubuntu default-fabric-image
