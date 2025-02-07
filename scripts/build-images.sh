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

if [[ -z "$DOCKERFILE" ]]; then
    echo "DOCKERFILE is not set"
    exit 1
fi

for spec in $IMAGE; do
    target=$(echo -n "$spec" | cut -d= -f1)
    imagetag=$(echo -n "$spec" | cut -d= -f2)
    docker buildx build \
        --platform "${PLATFORM// /,}" \
        --tag "$REGISTRY/$imagetag" \
        --target "$target" \
        --file "$DOCKERFILE" \
        .
done
