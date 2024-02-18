#!/bin/sh

name=fabric-ca           # Default Certificate Authority(CA) name.
suadmin=adminOrdererOrg0 # Default CA account name.
admin=adminOrdererOrg0   # Default org account name to be registered.
org=ordererOrganizations # Default organizations.
aff=ordererorg0          # Default affiliation of org to be registerd.
pw=ordererorg0password   # Default password of org to be registered.

help() {
  echo "registerOrgAdmin.sh [OPTIONS]"
  echo "                    -h           Explain options."
  echo "                    -n <string>  Setting CA docker container name."
  echo "                    -s <string>  Setting CA account name."
  echo "                    -a <string>  Setting name to be registerd for org admin."
  echo "                    -o <string>  Setting name of organization directory."
  echo "                    -f <string>  Setting name of admin's affiliation."
  echo "                    -p <string>  Setting password of org to be registered."
  exit 0
}

while getopts "n:s:a:o:f:p:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    s) suadmin=$OPTARG;;
    a) admin=$OPTARG;;
    o) org=$OPTARG;;
    f) aff=$OPTARG;;
    p) pw=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

# Change fabric-ca-client-config.yaml for registing org account.
docker cp org.cfg $suadmin:/root/testnet/
docker exec -itd $suadmin mv /root/testnet/fabric-ca-client-config.yaml /root/testnet/fabric-ca-client-config.back
docker exec -itd $suadmin bash -c "cat /root/testnet/fabric-ca-client-config.back | head -133 > /root/testnet/fabric-ca-client-config.yaml"
docker exec -itd $suadmin bash -c "sed 's/testname/$admin/g' /root/testnet/org.cfg > /root/testnet/org.back"
docker exec -itd $suadmin bash -c "sed 's/testorg/$aff/g' /root/testnet/org.back > /root/testnet/org.cfg"
docker exec -itd $suadmin bash -c "cat /root/testnet/org.cfg >> /root/testnet/fabric-ca-client-config.yaml"
docker exec -itd $suadmin bash -c "cat /root/testnet/fabric-ca-client-config.back | tail -28 >> /root/testnet/fabric-ca-client-config.yaml"
# Register
docker exec -it $suadmin bash -c "fabric-ca-client register --id.secret=$pw --tls.certfiles /root/testnet/tls-cert.pem"
sleep 1
#docker exec -itd $suadmin mv /root/testnet/fabric-ca-client-config.back /root/testnet/fabric-ca-client-config.yaml

# Generate
docker exec -itd $admin mkdir -p /root/testnet/crypto-config/$org/$aff/users/$admin
docker exec -it $admin fabric-ca-client enroll -u https://$admin:$pw@$name:7054 -H /root/testnet/crypto-config/$org/$aff/users/$admin --tls.certfiles /root/testnet/tls-cert.pem --csr.hosts $admin
sleep 1
#docker exec -it $admin tree /root/testnet
docker exec -itd $admin bash -c "mv /root/testnet/crypto-config/$org/$aff/users/$admin/msp/cacerts/* /root/testnet/crypto-config/$org/$aff/users/$admin/msp/cacerts/ca.crt"
docker exec -itd $admin bash -c "mv /root/testnet/crypto-config/$org/$aff/users/$admin/msp/keystore/* /root/testnet/crypto-config/$org/$aff/users/$admin/msp/keystore/server.key"
docker exec -itd $admin mkdir -p /root/testnet/crypto-config/$org/$aff/users/$admin/msp/admincerts
docker exec -itd $admin mkdir -p /root/testnet/crypto-config/$org/$aff/users/$admin/msp/tlscacerts
docker exec -itd $admin cp /root/testnet/crypto-config/$org/$aff/users/$admin/msp/signcerts/cert.pem /root/testnet/crypto-config/$org/$aff/users/$admin/msp/admincerts/${admin}-cert.pem
docker exec -itd $admin cp /root/testnet/crypto-config/$org/$aff/users/$admin/msp/cacerts/ca.crt /root/testnet/crypto-config/$org/$aff/users/$admin/msp/tlscacerts
docker cp $admin:/root/testnet/crypto-config/$org/$aff/users/$admin/msp/admincerts/${admin}-cert.pem .
docker cp $admin:/root/testnet/crypto-config/$org/$aff/users/$admin/msp/keystore/server.key ./${admin}-server.key
