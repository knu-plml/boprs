#!/bin/sh

name=fabric-ca           # Default Certificate Authority(CA) name.
admin=adminOrdererOrg0   # Default org admin name.
org=ordererOrganizations # Default organizations.
aff=ordererorg0          # Default affiliation of org to be registerd.
typ=orderer              # Default node type.
user=orderer0            # Default node(user) name.
pw=orderer0password      # Default password of org to be registered.

help() {
  echo "registerNode.sh [OPTIONS]"
  echo "                -h           Explain options."
  echo "                -n <string>  Setting CA docker container name."
  echo "                -a <string>  Setting name to be registerd for org admin."
  echo "                -o <string>  Setting name of organization directory."
  echo "                -f <string>  Setting name of admin's affiliation."
  echo "                -t <string>  Setting node type."
  echo "                -u <string>  Setting node name."
  echo "                -p <string>  Setting password of org to be registered."
  exit 0
}

while getopts "n:a:o:f:t:u:p:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    a) admin=$OPTARG;;
    o) org=$OPTARG;;
    f) aff=$OPTARG;;
    t) typ=$OPTARG;;
    u) user=$OPTARG;;
    p) pw=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

# Change fabric-ca-client-config.yaml for registing org account.
docker cp node.cfg $admin:/root/testnet/crypto-config/$org/$aff/users/$admin/
docker exec -itd $admin mv /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.yaml /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.back
docker exec -itd $admin bash -c "cat /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.back | head -133 > /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.yaml"
docker exec -itd $admin bash -c "sed 's/testname/$user/g' /root/testnet/crypto-config/$org/$aff/users/$admin/node.cfg > /root/testnet/crypto-config/$org/$aff/users/$admin/node.back"
docker exec -itd $admin bash -c "sed 's/testorg/$aff/g' /root/testnet/crypto-config/$org/$aff/users/$admin/node.back > /root/testnet/crypto-config/$org/$aff/users/$admin/node.cfg"
docker exec -itd $admin bash -c "sed 's/testnode/$typ/g' /root/testnet/crypto-config/$org/$aff/users/$admin/node.cfg > /root/testnet/crypto-config/$org/$aff/users/$admin/node.back"
docker exec -itd $admin bash -c "cat /root/testnet/crypto-config/$org/$aff/users/$admin/node.back >> /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.yaml"
docker exec -itd $admin bash -c "cat /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.back | tail -28 >> /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.yaml"

# Register
docker exec -it $admin bash -c "fabric-ca-client register --id.secret=$pw -H /root/testnet/crypto-config/$org/$aff/users/$admin/ --tls.certfiles /root/testnet/tls-cert.pem"
sleep 1
docker exec -itd $admin mv /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.back /root/testnet/crypto-config/$org/$aff/users/$admin/fabric-ca-client-config.yaml 

# Generate
docker exec -itd $user mkdir -p /root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/
docker exec -it $user bash -c "fabric-ca-client enroll -u https://$user:$pw@$name:7054 -H /root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/ --tls.certfiles /root/testnet/tls-cert.pem --csr.hosts $user"
sleep 1
docker exec -itd $user bash -c "mv /root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/msp/cacerts/* /root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/msp/cacerts/ca.crt"
docker exec -itd $user bash -c "mv /root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/msp/keystore/* /root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/msp/keystore/server.key"

docker exec -itd $user mkdir -p /root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/msp/admincerts/
docker cp ${admin}-cert.pem $user:/root/testnet/crypto-config/$org/$aff/${typ}s/${user}.${aff}/msp/admincerts
