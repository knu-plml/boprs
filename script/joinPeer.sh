#!/bin/sh

name=adminJistap # Default org admin name.
aff=jistap       # Default admin's affiliation.
msp=JISTAP    # Default admin's MSP
peer=peer0     # Default peer name.
ord=orderer0   # Default orderer name.

help() {
  echo "joinPeer.sh [OPTIONS]"
  echo "            -h           Explain options."
  echo "            -n <string>  Setting CA docker container name."
  echo "            -f <string>  Setting name of admin's affiliation."
  echo "            -m <string>  Setting name of admin's MSP."
  echo "            -p <string>  Setting peer name."
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
docker exec -it $name bash -c "CORE_PEER_LOCALMSPID=\"${msp}\" \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=true \\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/$aff/users/$name/msp/keystore/server.key \\
CORE_PEER_ADDRESS=$peer:7051 \\
peer channel join -b /root/testnet/boprs.block"
# --tls --cafile /root/testnet/${ord}-cert.pem"

