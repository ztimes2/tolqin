#!/bin/bash

REVISION=$(git rev-parse HEAD)
TAG=ztimes2/tolqin-api:$REVISION

docker build -t $TAG .
docker push $TAG