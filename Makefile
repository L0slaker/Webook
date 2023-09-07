.PHONY: docker
docker:
#	@rm webook || true
	@if exist webook del webook
	@go env -w GOOS=linux
	@go env -w GOARCH=amd64
	@go build -tags=k8s -o webook .
	@docker rmi -f l0slakers/webook:v0.0.1
	@docker build -t l0slakers/webook:v0.0.1 .