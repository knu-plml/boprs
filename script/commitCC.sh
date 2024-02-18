#!/bin/sh

name=${name}    # Default node name.
version=1   # Default msp name.
sequence=1    # Default couchdb docker container name.
has=0
package_id=0

help() {
  exit 0
}

while getopts "n:l:v:s:a:p:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    v) version=$OPTARG;;
    s) sequence=$OPTARG;;
    a) has=$OPTARG;;
    p) package_id=$OPTARG;;
    h) help;;
    ?) help ;;
  esac
done

echo name=${name}
echo version=${version}
echo sequence=${sequence}

if [ ${has} -ne 0 ] ; then
    echo hash=${has}
    package_id=${name}_${version}_${sequence}:${has}
fi

echo package_id=${package_id}


docker exec -it adminJistap bash -c "CORE_PEER_LOCALMSPID=\"JISTAP\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/keystore/server.key \\
CORE_PEER_ADDRESS=peer0:7051 \\
peer lifecycle chaincode approveformyorg --tls --cafile /root/testnet/orderer0-cert.pem -C boprs --name ${name} --peerAddresses peer0:7051 --tlsRootCertFiles /root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt  --version ${version}.${sequence} --package-id ${package_id} --sequence ${sequence} --waitForEvent"

docker exec -it adminJistap2 bash -c "CORE_PEER_LOCALMSPID=\"JISTAP2\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/keystore/server.key \\
CORE_PEER_ADDRESS=peer2:7051 \\
peer lifecycle chaincode approveformyorg --tls --cafile /root/testnet/orderer0-cert.pem -C boprs --name ${name} --peerAddresses peer2:7051 --tlsRootCertFiles /root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp/cacerts/ca.crt  --version ${version}.${sequence} --package-id ${package_id} --sequence ${sequence} --waitForEvent"

docker exec -it adminJistap bash -c "CORE_PEER_LOCALMSPID=\"JISTAP\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/keystore/server.key \\
CORE_PEER_ADDRESS=peer0:7051 \\
peer lifecycle chaincode commit -o orderer0:7050 --tls --cafile /root/testnet/orderer0-cert.pem --peerAddresses peer0:7051 --tlsRootCertFiles /root/testnet/crypto-config/peerOrganizations/jistap/msp/cacerts/ca.crt --peerAddresses peer2:7051 --tlsRootCertFiles /root/testnet/crypto-config/peerOrganizations/jistap/msp/cacerts/ca.crt --channelID boprs --name ${name} --version ${version}.${sequence} --sequence ${sequence}"

docker exec -it adminJistap bash -c "CORE_PEER_LOCALMSPID=\"JISTAP\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp/keystore/server.key \\
CORE_PEER_ADDRESS=peer0:7051 \\
peer lifecycle chaincode querycommitted --channelID boprs --name ${name} --tls --cafile /root/testnet/orderer0-cert.pem"

