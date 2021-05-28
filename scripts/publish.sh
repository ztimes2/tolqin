#!/bin/bash

REGISTRY=$1
REVISION=$(git rev-parse HEAD)
TAG=$REGISTRY:$REVISION 

docker build -t $TAG .
docker push $TAG