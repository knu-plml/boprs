#!/bin/sh

name=${name}    # Default node name.
version=1    # Default affiliation.
sequence=1

help() {
  exit 0
}

while getopts "n:v:s:h" opt
do
  case $opt in
    n) name=$OPTARG;;
    v) version=$OPTARG;;
    s) sequence=$OPTARG;;
    h) help ;;
    ?) help ;;
  esac
done

echo name=${name}
echo version=${version}
echo sequence=${sequence}

installCC_result=$(./installCC.sh -n ${name} -v ${version} -s ${sequence})
echo ${installCC_result}
package_id=$(echo ${installCC_result} | grep -o "${name}_${version}_${sequence}:[0-9a-z]*" | tail -1)
echo package_id=${package_id}
./commitCC.sh -n ${name} -v ${version} -s ${sequence} -p ${package_id}
