CORE_PEER_LOCALMSPID="JISTAP" \
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap/users/adminJistap/msp \
CORE_PEER_TLS_ENABLED=true \
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/msp/cacerts/ca.crt \
CORE_PEER_ADDRESS=peer0:7051 \
peer channel create -o orderer0:7050 -c boprs -f boprs.tx --tls --cafile orderer0-cert.pem
