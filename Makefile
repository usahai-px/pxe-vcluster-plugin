IMAGE_REP = docker.io/usahai728/px-sync-plugin
IMAGE_TAG = latest

VCLUSTER_NAME = pds


build: ## Build manager binary.
	go build -o bin/manager ./cmd/main.go

docker: docker-build docker-push

docker-build:
	docker build . -t ${IMAGE_REP}:${IMAGE_TAG}

docker-push:
	docker push ${IMAGE_REP}:${IMAGE_TAG}

vcluster:
	vcluster create ${VCLUSTER_NAME} -f ./vcluster.yaml --upgrade