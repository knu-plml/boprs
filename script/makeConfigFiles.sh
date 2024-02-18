#!/bin/sh

name=adminOrdererOrg0     # Default account name of CA.

help() {
  echo "makeConfigFiles.sh [OPTIONS]"
  echo "                   -h           Explain options."
  echo "                   -n <string>  Setting CA docker container name."
  exit 0
}

while getopts "n:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

# Hard coding...
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer0.ordererorg0/msp/signcerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer0.ordererorg0/msp/tlscacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer0.ordererorg0/msp/cacerts/
docker exec -itd $name bash -c "cp /root/testnet/msp/cacerts/* /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer0.ordererorg0/msp/tlscacerts/ca.crt"
docker cp orderer0:/root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer0.ordererorg0/msp/signcerts/cert.pem ./orderer0-cert.pem
docker cp ./orderer0-cert.pem $name:/root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer0.ordererorg0/msp/signcerts/

docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer1.ordererorg0/msp/signcerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer1.ordererorg0/msp/tlscacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer1.ordererorg0/msp/cacerts/
docker exec -itd $name bash -c "cp /root/testnet/msp/cacerts/* /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer1.ordererorg0/msp/tlscacerts/ca.crt"
docker cp orderer1:/root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer1.ordererorg0/msp/signcerts/cert.pem ./orderer1-cert.pem
docker cp ./orderer1-cert.pem $name:/root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer1.ordererorg0/msp/signcerts/

docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer2.ordererorg0/msp/signcerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer2.ordererorg0/msp/tlscacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer2.ordererorg0/msp/cacerts/
docker exec -itd $name bash -c "cp /root/testnet/msp/cacerts/* /root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer2.ordererorg0/msp/tlscacerts/ca.crt"
docker cp orderer2:/root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer2.ordererorg0/msp/signcerts/cert.pem ./orderer2-cert.pem
docker cp ./orderer2-cert.pem $name:/root/testnet/crypto-config/ordererOrganizations/ordererorg0/orderers/orderer2.ordererorg0/msp/signcerts/

docker exec -itd $name mkdir -p /root/testnet/crypto-config/peerOrganizations/jistap/msp/admincerts
docker exec -itd $name mkdir -p /root/testnet/crypto-config/peerOrganizations/jistap/msp/cacerts
docker exec -itd $name bash -c "cp /root/testnet/msp/cacerts/* /root/testnet/crypto-config/peerOrganizations/jistap/msp/cacerts/ca.crt"
docker cp ./adminJistap-cert.pem $name:/root/testnet/crypto-config/peerOrganizations/jistap/msp/admincerts

docker exec -itd $name mkdir -p /root/testnet/crypto-config/peerOrganizations/jistap2/msp/admincerts
docker exec -itd $name mkdir -p /root/testnet/crypto-config/peerOrganizations/jistap2/msp/cacerts
docker exec -itd $name bash -c "cp /root/testnet/msp/cacerts/* /root/testnet/crypto-config/peerOrganizations/jistap2/msp/cacerts/ca.crt"
docker cp ./adminJistap2-cert.pem $name:/root/testnet/crypto-config/peerOrganizations/jistap2/msp/admincerts

docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/ordererorg0/msp/admincerts
docker exec -itd $name bash -c "cp /root/testnet/crypto-config/ordererOrganizations/ordererorg0/users/$name/msp/admincerts/* /root/testnet/crypto-config/ordererOrganizations/ordererorg0/msp/admincerts"

docker cp ./configtx.yaml $name:/root/testnet
docker exec -it $name bash -c "configtxgen -profile TwoOrgsOrdererGenesis -outputBlock /root/testnet/genesis.block -channelID mychannel; configtxgen -profile TwoOrgsChannel -outputCreateChannelTx /root/testnet/boprs.tx -channelID boprs; configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate /root/testnet/JISTAPanchors.tx -channelID boprs -asOrg JISTAP; configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate /root/testnet/JISTAP2anchors.tx -channelID boprs -asOrg JISTAP2;"
sleep 5

docker cp $name:/root/testnet/genesis.block .
docker cp $name:/root/testnet/boprs.tx .
docker cp $name:/root/testnet/JISTAPanchors.tx .
docker cp $name:/root/testnet/JISTAP2anchors.tx .
docker cp $name:/root/testnet/crypto-config/peerOrganizations/jistap/msp/cacerts/ca.crt ./tlsca.crt
