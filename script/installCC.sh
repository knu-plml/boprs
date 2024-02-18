#!/bin/sh

name=${name}    # Default node name.
version=1    # Default affiliation.
sequence=1

help() {
  exit 0
}

while getopts "n:v:s:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    v) version=$OPTARG;;
    s) sequence=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

echo name=${name}
echo version=${version}
echo sequence=${sequence}

docker cp boprs adminJistap:/root/gopath/src/github.com/
docker exec -it adminJistap bash -c "CORE_PEER_LOCALMSPID=\"JISTAP\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true\\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/keystore/server.key \\
CORE_PEER_ADDRESS=peer0:7051 \\
peer lifecycle chaincode package /root/testnet/${name}.tar.gz --path github.com/boprs/${name} --lang golang --label ${name}_${version}_${sequence}"
docker cp adminJistap:/root/testnet/${name}.tar.gz ./
docker cp ./${name}.tar.gz adminJistap2:/root/testnet/

docker exec -it adminJistap bash -c "CORE_PEER_LOCALMSPID=\"JISTAP\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true\\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/keystore/server.key \\
CORE_PEER_ADDRESS=peer0:7051 \\
peer lifecycle chaincode install /root/testnet/${name}.tar.gz"

docker exec -it adminJistap2 bash -c "CORE_PEER_LOCALMSPID=\"JISTAP2\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/keystore/server.key \\
CORE_PEER_ADDRESS=peer2:7051 \\
peer lifecycle chaincode install /root/testnet/${name}.tar.gz"

wait

#docker exec -it adminJistap bash -c "CORE_PEER_LOCALMSPID=\"JISTAP\" \\
#CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp \\
#CORE_PEER_TLS_ENABLED=true \\
#CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
#CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
#CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
#CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/signcerts/cert.pem \\
#CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/keystore/server.key \\
#CORE_PEER_ADDRESS=peer0:7051 \\
#peer lifecycle chaincode queryinstalled"
#
#docker exec -it adminJistap2 bash -c "CORE_PEER_LOCALMSPID=\"JISTAP2\" \\
#CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp \\
#CORE_PEER_TLS_ENABLED=true \\
#CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/cacerts/ca.crt \\
#CORE_PEER_TLS_CLIENTAUTHREQUIRED=true\\
#CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/cacerts/ca.crt \\
#CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/signcerts/cert.pem \\
#CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/keystore/server.key \\
#CORE_PEER_ADDRESS=peer2:7051 \\
#peer lifecycle chaincode queryinstalled"
