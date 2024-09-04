#!/bin/sh

# Set RUNTIME_DIR
RUNTIME_DIR=${PWD}/runtime
RELEASE_PLEASE_CONFIG=${PWD}/release-please-config.json

# Go up until the root of the project (max 2 levels)
for _ in $(seq 1 2); do
  if [ -d $RUNTIME_DIR ]; then
    break
  fi
  cd ..
done
if ! [ -d $RUNTIME_DIR ]; then
  printf "Could not find %s\n", "$RUNTIME_DIR"
  exit 1
fi

# Set KYSTRAP_DIR
KYSTRAP_DIR=./tools/kystrap
if [ ! -d "$KYSTRAP_DIR" ]; then
  printf "Could not find %s\n", "$KYSTRAP_DIR"
  exit 1
fi

# check if -s flag is present anywhere in the arguments
# In that case we want to run in quiet mode and only output errors and results
SIMPLE_OUTPUT=false
for arg in "$@"
do
  case $arg in
    -s)
      SIMPLE_OUTPUT=true
      break
      ;;
  esac
done

# Check if -y flag is present anywhere in the arguments
# In that case we want to run in non-interactive mode
NON_INTERACTIVE=false
for arg in "$@"
do
  case $arg in
    -y)
      NON_INTERACTIVE=true
      break
      ;;
  esac
done

# Build docker image
if [ "$SIMPLE_OUTPUT" = true ]; then
  docker build --quiet --tag kystrap "$KYSTRAP_DIR" 1>/dev/null || exit 1
else
  docker build --tag kystrap "$KYSTRAP_DIR" || exit 1
fi

# Run docker image
if [ "$NON_INTERACTIVE" = true ]; then
  docker run \
    --rm                                                        `# Remove container after run` \
    --user "$(id -u):$(id -g)"                                  `# Run as current user` \
    --net="host"                                                `# Use host network` \
    --add-host=host.docker.internal:host-gateway                `# Add host.docker.internal to /etc/hosts` \
    -v "$RUNTIME_DIR":/app/runtime                              `# Mount runtime folder` \
    -v "$RELEASE_PLEASE_CONFIG":/app/release-please-config.json `# Mount release-please config` \
    kystrap $(echo "$@")                                         # Pass all arguments to kystrap
else
  docker run \
    -it                                                         `# Run in interactive mode` \
    --rm                                                        `# Remove container after run` \
    --user "$(id -u):$(id -g)"                                  `# Run as current user` \
    --net="host"                                                `# Use host network` \
    --add-host=host.docker.internal:host-gateway                `# Add host.docker.internal to /etc/hosts` \
    -v "$RUNTIME_DIR":/app/runtime                              `# Mount runtime folder` \
    -v "$RELEASE_PLEASE_CONFIG":/app/release-please-config.json `# Mount release-please config` \
    kystrap $(echo "$@")                                         # Pass all arguments to kystrap
fi