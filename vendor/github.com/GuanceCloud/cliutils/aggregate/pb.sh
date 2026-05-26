#!/bin/bash

GO_MODULE="github.com/GuanceCloud/cliutils"

# 1. Define Standard Imports
# We include the current directory (.)
INCLUDES="-I=. -I=.. -I=${GOPATH}/src -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf"

# 2. Define Mappings (The Critical Part)
# This tells gogoslick: "When you see an import for point/point.proto,
# use the Go package github.com/your-org/repo/point in the generated code."
# REPLACE 'github.com/your-org/repo' with your actual module name.
MAPPINGS="Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mpoint/point.proto=${GO_MODULE}/point"

# 3. Build aggrbatch.proto
echo "Building aggrbatch.proto & tsdata.proto"
protoc $INCLUDES \
    --gogoslick_out="${MAPPINGS}:." \
    aggregate/aggrbatch.proto \
    aggregate/tsdata.proto

if [ -f aggrbatch.pb.go ]; then
    mv aggrbatch.pb.go aggregate/aggrbatch.pb.go
fi

if [ -f tsdata.pb.go ]; then
    mv tsdata.pb.go aggregate/tsdata.pb.go
fi

echo "Done."
