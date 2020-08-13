go_path=../../../../../../../..
protoc ./common/*.proto --go_out=plugins=grpc:${go_path}
protoc ./language-agent-v2/*.proto --go_out=plugins=grpc:${go_path}
protoc ./register/*.proto --go_out=plugins=grpc:${go_path}