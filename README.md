# RAN DMS Operator
>
## Environments
### Hardware
- Ubuntu 18.04 (kernel 5.4.0-139) (VM from OpenStack)
- vCPU: 8 core (Intel(R) Core(TM) i9-10900 CPU @ 2.80GHz)
- RAM: 16G

### Software
- Go: go1.17.13 linux/amd64
- Operator-sdk: v1.17.0 (compatiable to kubernetes 1.21)

## Usage
Use `make` to perform operations:
### Install operator
```
make deploy
```

### Create OAI RAN Slice
* Deploy Slice
```
# config/samples place sample CRD
kubectl apply -f config/samples/ranslice_v1alpha1_oairanslice1.yaml
```

* Attach UE
```
# Attach to DU
kubectl exec -ti -n o-ran <DU-pods-name> -- /bin/bash

# Install requirements
apt install iputils-ping tmux -y

# Start up UE
cd cmake_targets/ran_build/build

## Switch into tmux session
tmux
RFSIMULATOR=127.0.0.1 ./nr-uesoftmodem -r 106 --numerology 1 --band 78 -C 3619200000 --rfsim --sa --nokrnmod -O ../../../targets/PROJECTS/GENERIC-NR-5GC/CONF/ue.conf

## Exit tmux session
<Ctrl+B> then D

# ping 
ping -I oaitun_ue1 8.8.8.8
```

## Build
>Check the image name at `Makefile` and `config/manager/kustomization.yaml`
```
# Build image at local
make docker-build

# Build and push image
make docker-build docker-push
```

## Support

## Known issue
### Failed to delete `oairanslice` crd
While you try to delete crd it will stuck:
```
$ k delete -f config/samples/ranslice_v1alpha1_oairanslice.yaml
oairanslice.ranslice.winlab.nycu "oairanslice-1" deleted
```
* Solution
Ctrl+C the `kubectl delete` command, and patch the finalizers field on crd:
```
# replace oairanslice.ranslice.winlab.nycu/oairanslice-1 with yor crd name, oairanslice.ranslice.winlab.nycu/<crd-name>
kubectl patch oairanslice.ranslice.winlab.nycu/oairanslice-1 -p '{"metadata":{"finalizers":[]}}' --type=merge
```
