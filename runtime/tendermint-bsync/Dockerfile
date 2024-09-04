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

# Build the docker core
RUN yarn run build:docker:binaries

################## Build Stage 2 ##################
# Create a slim runtime image with only the necessary files for execution
#################  Build Stage 2 #################
FROM node:slim AS runtime

# Set the port the container should expose
EXPOSE 50051

# Set the working directory
WORKDIR /usr/src/app

# Copy the compiled output from the build stage
COPY --from=build /usr/src/app/out/tendermint-bsync ./tendermint-bsync

# Set the command to run when the container starts
CMD ["./tendermint-bsync"]
