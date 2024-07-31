#!/bin/bash

# Define the repository URL and directories
REPO_URL="https://github.com/jito-labs/mev-protos"
REPO_DIR="mev-protos"
OUTPUT_DIR="../pb"
IMPORT_PATH="github.com/jito-labs/mev-protos/jito_pb"  # Adjust this to match your actual import path

# Fetch the repository
echo "Cloning repository..."
git clone $REPO_URL

# Check if the clone was successful
if [ ! -d "$REPO_DIR" ]; then
  echo "Failed to clone repository."
  exit 1
fi

# Change directory to the repository
cd $REPO_DIR

# Define the package mappings
PROTO_FILES=$(find . -name '*.proto')
MAPPING_ARGS=""

for file in $PROTO_FILES; do
  # Extract the file path relative to the repo root
  REL_PATH="${file#./}"
  # Add the mapping argument with the proper Go import path
  MAPPING_ARGS+="M${REL_PATH}=${IMPORT_PATH},"
done

# Create output directory if it doesn't exist
mkdir -p "../$OUTPUT_DIR"

# Generate Go code from protobuf definitions
echo "Generating Go code from protobuf definitions..."
protoc --proto_path=. --go_out="../$OUTPUT_DIR" --go_opt=paths=source_relative \
       --go-grpc_out="../$OUTPUT_DIR" --go-grpc_opt=paths=source_relative \
       --go_opt=${MAPPING_ARGS%,} --go-grpc_opt=${MAPPING_ARGS%,} \
       $(find . -name '*.proto')

# Check if the generation was successful
if [ $? -ne 0 ]; then
  echo "Failed to generate Go code from protobuf definitions."
  exit 1
fi

# Change back to the original directory
cd ..

# Clean up by removing the cloned repository
echo "Cleaning up..."
rm -rf $REPO_DIR

echo "Done. Generated files are in the '$OUTPUT_DIR' directory."
