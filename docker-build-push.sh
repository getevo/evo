#!/bin/bash
set -e

REGISTRY=""
SHA=staging #only for manual push
URL="${REGISTRY}/evo"
CURRENT_BUILD=${URL}:${SHA}

docker build -t ${CURRENT_BUILD} .
docker push ${CURRENT_BUILD}

echo "-------------------------------------------"
echo "Pushed:    ${CURRENT_BUILD}"
