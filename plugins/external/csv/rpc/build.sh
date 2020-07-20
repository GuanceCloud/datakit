cp ../../../../io/dk.proto .
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. dk.proto
rm dk.proto
