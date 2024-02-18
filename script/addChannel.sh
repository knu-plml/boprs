#!/bin/sh

name=adminJistap # Default Certificate Authority(CA) name.
aff=jistap       # Default admin's affiliation.
msp=JISTAP    # Default admin's MSP
peer=peer0     # Default peer name.
ord=orderer0   # Default orderer name.

help() {
  echo "addChannel.sh [OPTIONS]"
  echo "              -h           Explain options."
  echo "              -n <string>  Setting CA docker container name."
  echo "              -f <string>  Setting name of admin's affiliation."
  echo "              -m <string>  Setting name of admin's MSP."
  echo "              -p <string>  Setting peer name."
  echo "              -o <string>  Setting orderer name."
  exit 0
}

while getopts "n:f:m:p:o:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    f) aff=$OPTARG;;
    m) msp=$OPTARG;;
    p) peer=$OPTARG;;
    o) ord=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

# Hard coding...
docker cp ./boprs.tx $name:/root/testnet
docker cp ${ord}-cert.pem $name:/root/testnet

docker exec -it $name bash -c "CORE_PEER_LOCALMSPID=\"${msp}\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/cacerts/ca.crt \\
CORE_PEER_ADDRESS=$peer:7051 \\
peer channel create -o $ord:7050 -c boprs -f /root/testnet/boprs.tx --outputBlock /root/testnet/boprs.block --tls --cafile /root/testnet/${ord}-cert.pem"
sleep 2

docker cp $name:/root/testnet/boprs.block .
docker cp boprs.block adminJistap2:/root/testnet/boprs.block

#CORE_PEER_TLS_CERT_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/signcerts/cert.pem \\
#CORE_PEER_TLS_KEY_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/keystore/server.key \\
#CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
#CORE_PEER_TLS_CLIENTROOTCAS_FILES=[/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/cacerts/ca.crt,orderer0-cert.pem] \\
#CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/signcerts/cert.pem \\
#CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/keystore/server.key \\
