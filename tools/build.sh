#!/usr/bin/env bash

if [[ -z $DOCKER_DEFAULT_PLATFORM ]]; then
	echo "DOCKER_DEFAULT_PLATFORM is not specified. Use 'make docker-build'"
	exit 1
fi

if [[ -z $IMG ]]; then
	echo "IMG is not specified. Use 'make docker-build'"
	exit 1
fi

if [[ ${DOCKER_PUSH:-1} == 1 ]]; then
	imgresult="--push=true"
else
	imgresult="--load"
fi

if echo "$DOCKER_DEFAULT_PLATFORM" | grep -q ','; then
	if [ "${DOCKER_PUSH:-1}" = 0 ]; then
		echo "DOCKER_PUSH=0 option is not supported in case of multi-arch builds, please use DOCKER_PUSH=1"
		exit 1
	fi
fi

docker buildx build \
	--platform "$DOCKER_DEFAULT_PLATFORM" \
	--progress plain \
	--no-cache \
	$imgresult \
	-t "$IMG" \
	-f ./Dockerfile \
	.
