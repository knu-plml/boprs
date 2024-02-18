#!/bin/sh

name=fabric-ca         # Default Certificate Authority(CA) name.
admin=adminOrdererOrg0 # Default account name to be created of CA.
id=testadmin           # Default CA id.
pw=testadminpw         # Default CA password.

help() {
  echo "createCAadmin.sh [OPTIONS]"
  echo "                 -h           Explain options."
  echo "                 -n <string>  Setting CA docker container name."
  echo "                 -a <string>  Setting name to be created of CA."
  echo "                 -i <string>  Setting CA id."
  echo "                 -p <string>  Setting CA password."
  exit 0
}

while getopts "n:a:i:p:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    a) admin=$OPTARG;;
    i) id=$OPTARG;;
    p) pw=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

docker exec -it $admin bash -c "fabric-ca-client enroll -u https://$id:$pw@$name:7054 --tls.certfiles /root/testnet/tls-cert.pem --csr.cn $admin"

docker exec -it $admin bash -c "fabric-ca-client affiliation remove --force org1 --tls.certfiles /root/testnet/tls-cert.pem"
docker exec -it $admin bash -c "fabric-ca-client affiliation remove --force org2 --tls.certfiles /root/testnet/tls-cert.pem"

docker exec -it $admin bash -c "fabric-ca-client affiliation add jistap --tls.certfiles /root/testnet/tls-cert.pem"
docker exec -it $admin bash -c "fabric-ca-client affiliation add jistap2 --tls.certfiles /root/testnet/tls-cert.pem"
docker exec -it $admin bash -c "fabric-ca-client affiliation add ordererorg0 --tls.certfiles /root/testnet/tls-cert.pem"
