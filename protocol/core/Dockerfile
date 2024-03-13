################## Build Stage 1 ##################
# Create a staging Docker image with build dependencies and compile the app
#################  Build Stage 1 #################
FROM node:lts AS build

ARG VERSION

# Install jq
RUN apt-get update && apt-get install -y jq

# Set the working directory
WORKDIR /usr/src/app

# Copy necessary files for installing dependencies
COPY package.json tsconfig.json ./

# Update the version in the package.json file
RUN jq --arg VERSION "$VERSION" '.version=$VERSION' package.json > tmp.json && mv tmp.json package.json

# Install dependencies
RUN yarn install

# Copy source files
COPY ./src ./src

# Override the entrypoint to use the docker entrypoint
COPY ./src/entrypoint_docker.ts ./src/entrypoint.ts

# Build the docker core
RUN yarn run build:binaries

################## Build Stage 2 ##################
# Create a slim runtime image with only the necessary files for execution
#################  Build Stage 2 #################
FROM node:slim AS runtime

# Set the working directory
WORKDIR /usr/src/app

# Copy the compiled output from the build stage
COPY --from=build /usr/src/app/out/protocol ./

# Set the command to run when the container starts
CMD /bin/bash -c "set -a; source .env &>0; set +a; \
    ./protocol start \
    --pool \$POOL \
    --storage-priv \$STORAGE_PRIV \
    --request-backoff \$REQUEST_BACKOFF \
    --cache \$CACHE \
    --metrics \$METRICS \
    --metrics-port \$METRICS_PORT \
    --valaccount VALACCOUNT \
    --chain-id \$CHAIN_ID \
    --rpc \$RPC \
    --rest \$REST \
    \$DEBUG"
