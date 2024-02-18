#!/bin/sh

version=1    # Default affiliation.
sequence=1

help() {
  exit 0
}

while getopts "v:s:h" opt
do
  case $opt in
    v) version=$OPTARG;;
    s) sequence=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

echo version=${version}
echo sequence=${sequence}

docker cp boprs adminJistap:/root/gopath/src/github.com/
docker exec -it adminJistap bash -c "go env -w GO111MODULE=auto"
docker exec -w /root/gopath/src/github.com/boprs -it adminJistap bash -c "go mod init"
docker exec -w /root/gopath/src/github.com/boprs -it adminJistap bash -c "go mod tidy"
docker exec -w /root/gopath/src/github.com/boprs -it adminJistap bash -c "go mod vendor"

docker cp boprs adminJistap2:/root/gopath/src/github.com/
docker exec -it adminJistap2 bash -c "go env -w GO111MODULE=auto"
docker exec -w /root/gopath/src/github.com/boprs -it adminJistap2 bash -c "go mod init"
docker exec -w /root/gopath/src/github.com/boprs -it adminJistap2 bash -c "go mod tidy"
docker exec -w /root/gopath/src/github.com/boprs -it adminJistap2 bash -c "go mod vendor"

./instNcommCC.sh -n acceptanceModel -v ${version} -s ${sequence}
./instNcommCC.sh -n comment -v ${version} -s ${sequence}
./instNcommCC.sh -n contract -v ${version} -s ${sequence}
./instNcommCC.sh -n message -v ${version} -s ${sequence}
./instNcommCC.sh -n paper -v ${version} -s ${sequence}
./instNcommCC.sh -n rating -v ${version} -s ${sequence}
./instNcommCC.sh -n report -v ${version} -s ${sequence}
./instNcommCC.sh -n reviewer -v ${version} -s ${sequence}
