#!/bin/bash
set -eo pipefail
shopt -s nullglob
set -o xtrace

src_dir="$(realpath $(dirname $0)/..)"
git_branch=$(git rev-parse --abbrev-ref HEAD)

pushd ${src_dir}
image_hash=$(docker build -t perconalab/percona-version-service:$git_branch . | grep -E "^Successfully built .*" | awk '{print $3}' )
popd

echo "$image_hash"
docker tag "$image_hash" perconalab/percona-xtradb-cluster-operator:$git_branch
docker push perconalab/version-service:$git_branch

