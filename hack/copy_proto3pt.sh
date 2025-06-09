#!/bin/bash

# Function to copy .proto files from a source to a destination while preserving
# the directory structure
copy_proto_files() {
  local SRC_DIR="$1"
  local DST_DIR="$2"

  if [[ ! -d "$SRC_DIR" ]]; then
    echo "Source directory does not exist: $SRC_DIR"
    return
  fi

  # Create the root destination directory if it doesn't exist
  mkdir -p "$DST_DIR"

  # Traverse and process each .proto file
  find "$SRC_DIR" -type f -name "*.proto" | while read -r src_file; do
    # Calculate the relative path from the source directory
    local rel_path="${src_file#$SRC_DIR/}"

    # Construct the full destination file path
    local dst_file="$DST_DIR/$rel_path"
    local dst_dir
    dst_dir=$(dirname "$dst_file")

    # Create the destination subdirectory if it doesn't exist
    mkdir -p "$dst_dir"

    # Copy the file
    cp "$src_file" "$dst_file"
    echo "Copied: $src_file -> $dst_file"
  done
}

# Check if number of arguments is even
if (( $# % 2 != 0 )); then
  echo "Usage: $0 <src1> <dst1> [<src2> <dst2> ...]"
  exit 1
fi

copy_proto_files "./_patches/googleapis/google/api" \
"./api/third_party/googleapis/google/api"

copy_proto_files "./_patches/grpc-gateway/protoc-gen-openapiv2" \
"./api/third_party/grpc-gateway/protoc-gen-openapiv2"

copy_proto_files "./_patches/protovalidate/proto/protovalidate" \
"./api/third_party/protovalidate/proto/protovalidate"
