#!/bin/sh

name=fabric-ca           # Default Certificate Authority(CA) name.
admin=adminOrdererOrg0   # Default org account name to get CA cert.
org=ordererOrganizations # Default organizations.
aff=ordererorg0          # Default affiliation of org to be registerd.

help() {
  echo "getCAcertonOrgAdmin.sh [OPTIONS]"
  echo "                       -h           Explain options."
  echo "                       -n <string>  Setting CA docker container name."
  echo "                       -a <string>  Setting name to get CA certificate."
  echo "                       -f <string>  Setting name of admin's affiliation."
  echo "                       -o <string>  Setting name of organization directory."
  exit 0
}

while getopts "n:a:o:f:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    a) admin=$OPTARG;;
    o) org=$OPTARG;;
    f) aff=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

docker exec -itd $admin mkdir -p /root/testnet/crypto-config/$org/$aff/msp
docker exec -it $admin bash -c "fabric-ca-client getcacert -u https://$name:7054 -M /root/testnet/crypto-config/$org/$aff/msp --tls.certfiles tls-cert.pem"

# Change CA certificate name for using script easily.
docker exec -itd $admin bash -c "mv /root/testnet/crypto-config/$org/$aff/msp/cacerts/* /root/testnet/crypto-config/$org/$aff/msp/cacerts/ca.crt"

docker exec -itd $admin mkdir -p /root/testnet/crypto-config/$org/$aff/msp/tlscacerts
docker exec -itd $admin bash -c "cp /root/testnet/crypto-config/$org/$aff/msp/cacerts/ca.crt /root/testnet/crypto-config/$org/$aff/msp/tlscacerts"

