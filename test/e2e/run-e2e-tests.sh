#!/bin/sh

E2ETEST_DIR=${PWD}/test/e2e

# Capture the first argument which will be passed to the e2e test
arg="$1"

echo "Running e2e tests with args: $arg"

# Go up until the root of the project (max 3 levels)
for _ in $(seq 1 3); do
  if [ -d $E2ETEST_DIR ]; then
    break
  fi
  cd ..
done
if ! [ -d $E2ETEST_DIR ]; then
  printf "Could not find %s\n", "$E2ETEST_DIR"
  exit 1
fi

# Build e2etest docker image
docker build $E2ETEST_DIR -t e2etest

# Run docker image in privileged mode (required for docker-in-docker)
# Use this only for local testing!!!
docker run \
    --rm                                     `# Remove container after run` \
    --privileged                             `# Run in privileged mode` \
    -d                                       `# Run in detached mode` \
    --name e2etest                           `# Name the container` \
    -v ${PWD}/runtime:/mnt/runtime:ro        `# Mount runtime folder` \
    -v ${PWD}/tools:/tools:ro                `# Mount tools folder` \
    -v ${PWD}/protocol:/protocol:ro          `# Mount protocol folder` \
    -v e2etestvol:/var/lib/docker            `# Mount docker volume (to cache images)` \
    e2etest

# Check if the docker daemon inside the container is running (it takes a few seconds to start)
for _ in $(seq 1 10); do
  if docker exec -it e2etest sh -c "docker ps"; then
    break
  fi
  sleep 1
done

# Run e2e tests in docker container
# Make a copy of runtime folder in the container
docker exec -it e2etest sh -c \
  "rm -rf /runtime;
  mkdir /runtime;
  cp -r /mnt/runtime/* /runtime;
  ./e2etest -test.v -test.parallel 10 -test.timeout 30m $arg" || docker stop e2etest; echo "e2e tests failed!"; exit 1

# Stop docker container
docker stop e2etest