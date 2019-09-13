#! /bin/bash
TAG=$1
echo "Tag: $TAG"

if [[ $TAG == "" ]]; then
  echo "missing TAG"
  exit 1
fi

docker build -f ./docker/dnsupdate/Dockerfile -t rverst/dnsupdate:"$TAG" .
docker tag rverst/dnsupdate:"$TAG" rverst/dnsupdate:latest
