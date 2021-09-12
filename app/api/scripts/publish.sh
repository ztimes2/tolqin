#!/bin/bash

REGISTRY=$1
if [[ $REGISTRY == '' ]]; then
    echo "missing argument: must specify a Docker image repository"
    exit 1
fi

REVISION=$(git rev-parse HEAD)
TAG=$REGISTRY:$REVISION 

docker build -t $TAG .
docker push $TAG