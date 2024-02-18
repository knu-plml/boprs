#!/bin/sh

name=orderer0     # Default Certificate Authority(CA) name.
aff=ordererorg0   # Default affiliation.

help() {
  echo "startOrderer.sh [OPTIONS]"
  echo "                -h           Explain options."
  echo "                -n <string>  Setting CA docker container name."
  echo "                -f <string>  Setting name of orderer's affiliation."
  exit 0
}

while getopts "n:f:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    f) aff=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

# Hard coding...
docker cp ./genesis.block $name:/root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/
docker cp ./configtx.yaml $name:/root/testnet

docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer0.${aff}/msp/cacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer0.${aff}/msp/tlscacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer0.${aff}/msp/signcerts/
docker cp ./orderer0-cert.pem $name:/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/orderer0.${aff}/msp/signcerts/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer0.${aff}/msp/cacerts/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer0.${aff}/msp/tlscacerts/

docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer1.${aff}/msp/cacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer1.${aff}/msp/tlscacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer1.${aff}/msp/signcerts/
docker cp ./orderer1-cert.pem $name:/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/orderer1.${aff}/msp/signcerts/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer1.${aff}/msp/cacerts/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer1.${aff}/msp/tlscacerts/

docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer2.${aff}/msp/cacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer2.${aff}/msp/tlscacerts/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer2.${aff}/msp/signcerts/
docker cp ./orderer2-cert.pem $name:/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/orderer2.${aff}/msp/signcerts/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer2.${aff}/msp/cacerts/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/orderer2.${aff}/msp/tlscacerts/

docker exec -itd $name rm -f /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/signcerts/cert.pem

docker exec -itd $name mkdir -p /root/testnet/crypto-config/peerOrganizations/org0/peers/peer0.org0/tls/
docker exec -itd $name mkdir -p /root/testnet/crypto-config/peerOrganizations/org1/peers/peer2.org1/tls/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/peerOrganizations/org0/peers/peer0.org0/tls/
docker exec -itd $name cp /root/testnet/crypto-config/ordererOrganizations/$aff/orderers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/peerOrganizations/org1/peers/peer2.org1/tls/

docker exec -it $name bash -c "FABRIC_LOGGING_SPEC=INFO \\
CORE_CHAINCODE_LOGGING_LEVEL=DEBUG \\
CORE_VM_DOCKER_ATTACHSTDOUT=true \\
ORDERER_GENERAL_LISTENADDRESS=${name} \\
ORDERER_GENERAL_GENESISMETHOD=file \\
ORDERER_GENERAL_GENESISFILE=/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/genesis.block \\
ORDERER_GENERAL_LOCALMSPID=OrdererOrg0MSP \\
ORDERER_GENERAL_LOCALMSPDIR=/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp \\
ORDERER_GENERAL_TLS_CLIENTAUTHREQUIRED=false \\
ORDERER_GENERAL_TLS_ENABLED=true \\
ORDERER_GENERAL_TLS_CERTIFICATE=/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/signcerts/${name}-cert.pem \\
ORDERER_GENERAL_TLS_PRIVATEKEY=/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/keystore/server.key \\
ORDERER_GENERAL_TLS_ROOTCAS=[/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/cacerts/ca.crt] \\
ORDERER_GENERAL_CLUSTER_ROOTCAS=[/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/cacerts/ca.crt] \\
ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY=/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/keystore/server.key \\
ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE=/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/signcerts/${name}-cert.pem \\
orderer"

#ORDERER_GENERAL_TLS_ROOTCAS=[/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/cacerts/ca.crt,/root/testnet/crypto-config/peerOrganizations/org0/peers/peer0.org0/tls/ca.crt,/root/testnet/crypto-config/peerOrganizations/org1/peers/peer2.org1/tls/ca.crt] \\
#ORDERER_GENERAL_CLUSTER_ROOTCAS=[/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/cacerts/ca.crt,/root/testnet/crypto-config/peerOrganizations/org0/peers/peer0.org0/tls/ca.crt,/root/testnet/crypto-config/peerOrganizations/org1/peers/peer2.org1/tls/ca.crt] \\
#ORDERER_GENERAL_TLS_CLIENTAUTHREQUIRED=true \\
#ORDERER_GENERAL_TLS_CLIENTROOTCAS=/root/testnet/crypto-config/ordererOrganizations/${aff}/orderers/${name}.${aff}/msp/cacerts/ca.crt \\
