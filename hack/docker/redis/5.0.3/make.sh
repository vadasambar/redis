#!/bin/bash
set -xeou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/kubedb.dev/redis

source "$REPO_ROOT/hack/libbuild/common/lib.sh"
source "$REPO_ROOT/hack/libbuild/common/kubedb_image.sh"

DOCKER_REGISTRY=${DOCKER_REGISTRY:-kubedb}
IMG=redis
SUFFIX=v1
TAG="5.0.3-$SUFFIX"
DIR=5.0.3

build() {
  pushd "$REPO_ROOT/hack/docker/redis/$DIR"

  local cmd="docker build -t $DOCKER_REGISTRY/$IMG:$TAG ."
  echo $cmd; $cmd

  popd
}

push() {
  pushd "$REPO_ROOT/hack/docker/redis/$DIR"

  local cmd="docker push $DOCKER_REGISTRY/$IMG:$TAG"
  echo $cmd; $cmd

  popd
}

binary_repo $@
