#!/bin/sh

name=fabric-ca     # Default Certificate Authority(CA) name.
id=testadmin       # Default CA id.
pw=testadminpw     # Default CA password.

help() {
  echo "startCA.sh [OPTIONS]"
  echo "           -h           Explain options."
  echo "           -n <string>  Setting CA docker container name."
  echo "           -i <string>  Setting CA id."
  echo "           -p <string>  Setting CA password."
  exit 0
}

while getopts "n:i:p:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    i) id=$OPTARG;;
    p) pw=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

# I don't know why tls-cert.pem has to be created and removed.
# But fabric server isn't work, if i don't do that.
docker exec -itd $name touch /root/testnet/tls-cert.pem
docker exec -itd $name rm /root/testnet/tls-cert.pem
docker exec -itd $name fabric-ca-server start -b $id:$pw --tls.enabled --cfg.affiliations.allowremove --cfg.identities.allowremove -d --csr.hosts $name
# Time of turning on.
sleep 2
docker cp $name:/root/testnet/tls-cert.pem .
