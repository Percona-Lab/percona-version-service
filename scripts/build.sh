#!/bin/bash
set -eo pipefail
shopt -s nullglob
set -o xtrace

src_dir="$(realpath $(dirname $0)/..)"
git_branch=$(git rev-parse --abbrev-ref HEAD)
image=perconalab/version-service:$git_branch 
pushd ${src_dir}
docker build -t $image .
popd

docker push $image 

