#!/bin/sh

net=testNetwork # Default docker network name.

#./makeDefaultImage.sh

#Load docker image from docker image file.
docker load < default-fabric-image.tar

#Create docker network for testing fabric network.
./createDockerNetwork.sh
#./createDockerNetwork.sh

#Create docker nodes for testing fabric network.
./createNode.sh -c fabric-ca -n $net -p "10054:7054"
./startCA.sh

docker pull couchdb:2.3

./createNode.sh -c adminJistap -n $net
./createNodeDooD.sh -c peer0 -n $net -p "10501:7051" -e "11501:7053"
./createCouch.sh -c couchdb0 -n $net -i couchdb:2.3
./createNodeDooD.sh -c peer1 -n $net -p "10502:7051"
./createCouch.sh -c couchdb1 -n $net -i couchdb:2.3

./createNode.sh -c adminJistap2 -n $net
./createNodeDooD.sh -c peer2 -n $net -p "10503:7051" -e "11503:7053"
./createCouch.sh -c couchdb2 -n $net -i couchdb:2.3
./createNodeDooD.sh -c peer3 -n $net -p "10504:7051"
./createCouch.sh -c couchdb3 -n $net -i couchdb:2.3

./createNode.sh -c adminOrdererOrg0 -n $net
./createNode.sh -c orderer0 -n $net -p "10510:7050"
./createNode.sh -c orderer1 -n $net -p "10520:7050"
./createNode.sh -c orderer2 -n $net -p "10530:7050"

./createCAadmin.sh
echo ""

echo "getCAcertonOrgAdmin"
./getCAcertonOrgAdmin.sh -a adminOrdererOrg0 -o ordererOrganizations -f ordererorg0
./getCAcertonOrgAdmin.sh -a adminJistap -o peerOrganizations -f jistap
./getCAcertonOrgAdmin.sh -a adminJistap2 -o peerOrganizations -f jistap2
echo ""

echo "registerOrgAdmin"
./registerOrgAdmin.sh -a adminOrdererOrg0 -o ordererOrganizations -f ordererorg0 -p ordererorg0password
./registerOrgAdmin.sh -a adminJistap -o peerOrganizations -f jistap -p jistappassword
./registerOrgAdmin.sh -a adminJistap2 -o peerOrganizations -f jistap2 -p jistap2password
echo ""

echo "registerNode ordererorg0"
./registerNode.sh -a adminOrdererOrg0 -o ordererOrganizations -f ordererorg0 -t orderer -u orderer0 -p orderer0passwd
./registerNode.sh -a adminOrdererOrg0 -o ordererOrganizations -f ordererorg0 -t orderer -u orderer1 -p orderer1passwd
./registerNode.sh -a adminOrdererOrg0 -o ordererOrganizations -f ordererorg0 -t orderer -u orderer2 -p orderer2passwd
echo ""

echo "registerNode jistap"
./registerNode.sh -a adminJistap -o peerOrganizations -f jistap -t peer -u peer0 -p peer0passwd
./registerNode.sh -a adminJistap -o peerOrganizations -f jistap -t peer -u peer1 -p peer1passwd
echo ""

echo "registerNode jistap2"
./registerNode.sh -a adminJistap2 -o peerOrganizations -f jistap2 -t peer -u peer2 -p peer2passwd
./registerNode.sh -a adminJistap2 -o peerOrganizations -f jistap2 -t peer -u peer3 -p peer3passwd
echo ""

echo "makeConfigFiles"
./makeConfigFiles.sh -n adminOrdererOrg0
echo ""

echo "execute couchdb"
docker exec -itd couchdb0 tini -- /docker-entrypoint.sh /opt/couchdb/bin/couchdb
docker exec -itd couchdb1 tini -- /docker-entrypoint.sh /opt/couchdb/bin/couchdb
docker exec -itd couchdb2 tini -- /docker-entrypoint.sh /opt/couchdb/bin/couchdb
docker exec -itd couchdb3 tini -- /docker-entrypoint.sh /opt/couchdb/bin/couchdb
sleep 2
echo""

echo "startOrderer"
./startOrderer.sh -n orderer0 -f ordererorg0
./startOrderer.sh -n orderer1 -f ordererorg0
./startOrderer.sh -n orderer2 -f ordererorg0
echo ""

echo "startPeer"
./startPeer.sh -n peer0 -f jistap -m JISTAP -c couchdb0
./startPeer.sh -n peer1 -f jistap -m JISTAP -c couchdb1
./startPeer.sh -n peer2 -f jistap2 -m JISTAP2 -c couchdb2
./startPeer.sh -n peer3 -f jistap2 -m JISTAP2 -c couchdb3
echo ""

echo "addChannel"
./addChannel.sh -n adminJistap -f jistap -m JISTAP -p peer0 -o orderer0
echo ""

echo "joinPeer"
./joinPeer.sh -n adminJistap -f jistap -m JISTAP -p peer0
./joinPeer.sh -n adminJistap -f jistap -m JISTAP -p peer1
./joinPeer.sh -n adminJistap2 -f jistap2 -m JISTAP2 -p peer2
./joinPeer.sh -n adminJistap2 -f jistap2 -m JISTAP2 -p peer3
echo ""

echo "updateAnchorPeer"
./updateAnchorPeer.sh -n adminJistap -f jistap -m JISTAP -p peer0 -o orderer0
./updateAnchorPeer.sh -n adminJistap2 -f jistap2 -m JISTAP2 -p peer2 -o orderer0

#./rmTempFiles.sh
