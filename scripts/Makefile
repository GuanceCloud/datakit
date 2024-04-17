.PHONY: all

all:
	GOOS=linux GOARCH=amd64 go build -o v1-write-linux-amd64.out v1-write.go
	GOOS=linux GOARCH=amd64 go build -o v1-data-linux-amd64.out  v1-data.go 
	#sudo docker buildx build --platform linux/arm64,linux/amd64 -t registry.jiagouyun.com/datakit-tools/v1-write:v0.01 -f v1-write.dockerfile .
	#sudo docker build --platform linux/arm64 -t registry.jiagouyun.com/datakit-tools/v1-write:v0.0.1 -f v1-write.dockerfile .

# deplpy v1-write as k8s ervice base on v1-write.yaml
dep:
	kubectl apply -f v1-write.yaml # create deployment
	kubectl expose deployment v1-write --type=NodePort --name=v1-write-service # export service base on the deployment
	kubectl describe services v1-write-service # show service info

clean:
	kubectl delete deployment v1-write
	kubectl delete service v1-write-service
