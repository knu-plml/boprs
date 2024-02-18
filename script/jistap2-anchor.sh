CORE_PEER_LOCALMSPID="JISTAP2" \
CORE_PEER_MSPCONFIGPATH=/root/testnet/crypto-config/peerOrganizations/jistap2/users/adminJistap2/msp \
CORE_PEER_TLS_ROOTCERT_FILE=/root/testnet/crypto-config/peerOrganizations/jistap/msp/cacerts/ca.crt \
CORE_PEER_TLS_ENABLED=true \
CORE_PEER_ADDRESS=peer2:7051 \
peer channel create -o orderer0:7050 -c boprs -f JISTAP2anchors.tx --tls --cafile orderer0-cert.pem
