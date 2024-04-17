.PHONY: gogopb protobuf

gogopb:
	protoc \
		-I=${GOPATH}/src -I=${GOPATH}/src/github.com/gogo/protobuf/protobuf -I. \
		--gogoslick_out=Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:. point.proto 

protobuf:
	protoc --go_out=. point.proto
	#rm -rf $subdir
	#
	## Python
	#mkdir -p $subdir/python && protoc --python_out=$subdir/python *.proto
	## Java
	#mkdir -p $subdir/java && protoc --java_out=$subdir/java *.proto
	## ObjC
	#mkdir -p $subdir/objc && protoc --objc_out=$subdir/objc *.proto
	## PHP
	#mkdir -p $subdir/php && protoc --php_out=$subdir/php *.proto
	## C++
	#mkdir -p $subdir/cpp && protoc --cpp_out=$subdir/cpp *.proto
	## C#
	#mkdir -p $subdir/csharp && protoc --csharp_out=$subdir/csharp *.proto
	## Dart
	##mkdir -p $subdir/dart && protoc --dart_out=$subdir/dart *.proto

clean:
	rm -rf point.pb.go
