#!/usr/bin/env bash

set -e

if [[ -z "$IMAGE" ]]; then
    echo "IMAGE is not set"
    exit 1
fi

if [[ -z "$REGISTRY" ]]; then
    echo "REGISTRY is not set"
    exit 1
fi

if [[ -z "$PLATFORM" ]]; then
    echo "PLATFORM is not set"
    exit 1
fi

copy_platform() {
    local platform=$1
    local registry=$2
    local imagetag=$3
    local image
    local tag
    local suffix

    image=$(echo -n "$imagetag" | cut -d: -f1)
    tag=$(echo -n "$imagetag" | cut -d: -f2)
    suffix=$(echo -n "$platform" | cut -d/ -f2- | tr -d /)

    crane copy \
        --platform "$platform" \
        "$registry/$imagetag" \
        "$registry/$image/$suffix:$tag"
}

for spec in $IMAGE; do
    imagetag=$(echo -n "$spec" | cut -d= -f2)
    docker push "$REGISTRY/$imagetag"
    for platform in $PLATFORM; do
        copy_platform "$platform" "$REGISTRY" "$imagetag"
    done
done
