#protoc -I grpc --python_out=./grpc grpc/dk.proto

python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. dk.proto
