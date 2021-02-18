
# Image URL to use all building/pushing image targets
IMG ?= yamajik/kess:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"
# Mainfests build dist path
BUILD_DIST?=dist/kess.yaml

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

build: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	mkdir -p $(shell dirname ${BUILD_DIST})
	$(KUSTOMIZE) build config/default -o ${BUILD_DIST}

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

images: docker-build docker-push

k3d-cluster:
	k3d cluster create kess --api-port 6443 --no-lb -p 8000:80@server[0] --wait

k3d-import-base:
	k3d image import gcr.io/kubebuilder/kube-rbac-proxy:v0.5.0 -c kess

k3d-import:
	k3d image import ${IMG} -c kess

dev: dev-deploy

dev-deploy: docker-build k3d-import deploy

dev-logs:
	kubectl -n kess-system logs -l control-plane=controller-manager --all-containers -f

dev-samples:
	kubectl create ns kess-samples || echo "OK"
	kubectl -n kess-samples apply -f config/samples

dev-samples-list:
	kubectl -n kess-samples get kess,all,cm,ing

dev-samples-watch:
	watch -n 1 kubectl -n kess-samples get kess -o wide

dev-samples-watch-others:
	watch -n 1 kubectl -n kess-samples get all,cm -o wide

dev-samples-delete:
	kubectl -n kess-samples delete kess --all
