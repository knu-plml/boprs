#!/bin/sh

name=peer0     # Default node name.
aff=jistap       # Default affiliation.
msp=JISTAP    # Default msp name.
couch=couchdb0 # Default couchdb docker container name.

help() {
  echo "startPeer.sh [OPTIONS]"
  echo "               -h           Explain options."
  echo "               -n <string>  Setting CA docker container name."
  echo "               -f <string>  Setting name of orderer's affiliation."
  echo "               -m <string>  Setting MSP name."
  echo "               -c <string>  Setting couchdb name to connect with peer."
  exit 0
}

while getopts "n:f:m:c:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    f) aff=$OPTARG;;
    m) msp=$OPTARG;;
    c) couch=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

# Hard coding...
docker exec -itd $name mkdir -p /root/testnet/crypto-config/peerOrganizations/$aff/peers/${name}.${aff}/msp/tlscacerts
docker exec -itd $name bash -c "cp /root/testnet/crypto-config/peerOrganizations/$aff/peers/${name}.${aff}/msp/cacerts/ca.crt /root/testnet/crypto-config/peerOrganizations/$aff/peers/${name}.${aff}/msp/tlscacerts"

docker exec -it $name bash -c "CORE_PEER_ENDORSER_ENABLED=true \\
CORE_PEER_ADDRESS=${name}:7051 \\
CORE_PEER_CHAINCODELISTENADDRESS=${name}:7052 \\
CORE_PEER_ID=${name} \\
CORE_PEER_LOCALMSPID=${msp} \\
CORE_PEER_GOSSIP_EXTERNALENDPOINT=${name}:7051 \\
CORE_PEER_GOSSIP_USELEADERELECTION=true \\
CORE_PEER_GOSSIP_ORGLEADER=false \\
CORE_PEER_TLS_ENABLED=true \\
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/${aff}/peers/${name}.${aff}/msp/cacerts/ca.crt \\
CORE_PEER_TLS_KEY_FILE=/root/testnet/crypto-config/peerOrganizations/${aff}/peers/${name}.${aff}/msp/keystore/server.key \\
CORE_PEER_TLS_CERT_FILE=/root/testnet/crypto-config/peerOrganizations/${aff}/peers/${name}.${aff}/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTAUTHREQUIRED=false \\
CORE_PEER_TLS_CLIENTROOTCAS_FILES=/root/testnet/crypto-config/peerOrganizations/${aff}/peers/${name}.${aff}/msp/cacerts/ca.crt \\
CORE_PEER_TLS_CLIENTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/${aff}/peers/${name}.${aff}/msp/signcerts/cert.pem \\
CORE_PEER_TLS_CLIENTKEY_FILE=/root/testnet/crypto-config/peerOrganizations/${aff}/peers/${name}.${aff}/msp/keystore/server.key \\
CORE_PEER_TLS_SERVERHOSTOVERRIDE=${name} \\
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/${aff}/peers/${name}.${aff}/msp \\
CORE_LEDGER_STATE_STATEDATABASE=CouchDB \\
CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS=${couch}:5984 \\
peer node start"

