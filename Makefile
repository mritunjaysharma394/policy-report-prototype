all: docker

build: fmt vet
	GOOS=linux go build -o policyreport .

fmt:
	go fmt ./...

vet:
	go vet ./...

docker: build
	docker build . -t wg-policy/policyreport

codegen:
	./hack/update-codegen.sh