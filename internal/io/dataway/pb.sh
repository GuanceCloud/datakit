protoc \
	-I=${GOPATH}/src -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf -I. \
	--gogoslick_out=Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:. cachedata.proto
