# venafi-test-wizard
---
<!-- markdown-link-check-disable -->
[![LICENSE](https://img.shields.io/github/license/pingcap/chaos-mesh.svg)](https://github.com/kxk-4498/Venafi-test-wizard/blob/main/LICENSE)
[![Language](https://img.shields.io/badge/Language-Go-blue.svg)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/chaos-mesh/chaos-mesh)](https://goreportcard.com/report/github.com/kxk-4498/Venafi-test-wizard)
[![CodeQL](https://github.com/guilhem/freeipa-issuer/workflows/CodeQL/badge.svg)](https://github.com/kxk-4498/Venafi-test-wizard/actions?query=workflow%3ACodeQL)
![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkxk-4498%2FVenafi-test-wizard.svg?type=small)
<!-- markdown-link-check-enable -->
Testing tool for cert-manager in Kubernetes.

## Description
The vision for this project is to be the standard open-source testing tool for cert-manager deployments adopted by the Kubernetes community.

Please go to [Chaos Issuer Wiki](https://github.com/kxk-4498/Venafi-test-wizard/wiki) for further information.


## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

You will need the following command line tools installed on your PATH:

* [Git](https://git-scm.com/)
* [Golang v1.17+](https://golang.org/)
* [Docker v17.03+](https://docs.docker.com/install/)
* [Kind v0.9.0+](https://kind.sigs.k8s.io/docs/user/quick-start/)
* [Kubectl v1.11.3+](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [Kubebuilder v2.3.1+](https://book.kubebuilder.io/quick-start.html#installation)
* [Kustomize v3.8.1+](https://kustomize.io/)
* [lolcat](https://github.com/busyloop/lolcat)

You may also want to read: the [Kubebuilder Book](https://book.kubebuilder.io/) and the [cert-manager Concepts Documentation](https://cert-manager.io/docs/concepts/) for further background
information.

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/venafi-test-wizard:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/venafi-test-wizard:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

# Running cert-manager locally with very less certificate duration modification #
**Note:** We hold no intellectual property of cert-manager resources and are merely using a modified local version for manual testing of our testing wizard.
**Resources:** [Building cert-manager](https://cert-manager.io/docs/contributing/building/ ), [Developing with Kind](https://cert-manager.io/docs/contributing/kind/)

1. create a folder outside this github directory

```sh
mkdir cert-manager-local
cd cert-manager-local
```

2. git clone cert-manager

```sh
git clone https://github.com/cert-manager/cert-manager.git
```

3. Make sure you have all dependencies installed which are git, curl, GNU make, jq, docker and go.

4. go inside cert-manager git repository you cloned.
```sh
cd cert-manager
```

5. For modifying the certificate duration and renew before time:
- go inside cert-manager/internal/apis/certmanager/validation/certificate.go
- import "time" module in certificate.go
- replace the ValidateDuration function to the one given below.
- the function below changes the minimum cert duration of cert-manager to 4 minutes and mininum certificate renewal time to 2 minutes.

```sh
func ValidateDuration(crt *internalcmapi.CertificateSpec, fldPath *field.Path) field.ErrorList {
    el := field.ErrorList{}
    duration := util.DefaultCertDuration(crt.Duration)
    MinimumCertificateDuration := time.Minute * 4
    MinimumRenewBefore := time.Minute * 2
    if duration < MinimumCertificateDuration { //was cmapi.MinimumCertificateDuration
        el = append(el, field.Invalid(fldPath.Child("duration"), duration, fmt.Sprintf("certificate duration must be greater than %s", MinimumCertificateDuration))) // was  cmapi.MinimumCertificateDuration
    }
    // If spec.renewBefore is set, check that it is not less than the minimum.
    if crt.RenewBefore != nil && crt.RenewBefore.Duration < MinimumRenewBefore { //was cmapi.MinimumRenewBefore
        el = append(el, field.Invalid(fldPath.Child("renewBefore"), crt.RenewBefore.Duration, fmt.Sprintf("certificate renewBefore must be greater than %s", MinimumRenewBefore))) // was cmapi.MinimumRenewBefore
    }
    // If spec.renewBefore is set, it must be less than the duration.
    if crt.RenewBefore != nil && crt.RenewBefore.Duration >= duration {
        el = append(el, field.Invalid(fldPath.Child("renewBefore"), crt.RenewBefore.Duration, fmt.Sprintf("certificate duration %s must be greater than renewBefore %s", duration, crt.RenewBefore.Duration)))
    }
    return el
}
```

6. Run cert-manager using this command (our Chaos Issuer runs on K8S_VERSION=1.25). This command will deploy a kind cluster nameed kind and deploy cert-manager resources in it:

 ```sh
 make K8S_VERSION=1.25 e2e-setup-kind e2e-setup-certmanager
 ```
**Note:** Running cert-manager locally takes 3-5 minutes to successfully deploy.

7. Deploy cert-manager CRDs
```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.crds.yaml
```

# Test It Out using Script #
**NOTE:** Make sure you enable kubectl to port forward for ports 80 & 443 by (For more info: [stackoverflow](https://stackoverflow.com/questions/53775328/kubernetes-port-forwarding-error-listen-tcp4-127-0-0-188-bind-permission-de)):
```
sudo setcap CAP_NET_BIND_SERVICE=+eip /usr/bin/kubectl (or /usr/local/bin/kubectl. Depends on where it is installed)
```
Incase of line ending problems if using on Mac or WSL2:
```
sed -i -e 's/\r$//' chaos_script.sh
```
0. Run the script

```sh
.\chaos_script.sh
```

1. Create the environment with the dependencies for chaos-testing:

```sh
chaos setup
```

2. Deploy the chaos issuer:

```sh
chaos deploy issuer
```

3. Deploy the certificate:

```sh
chaos deploy cert
```

4. To see available certificates:

```sh
chaos show cert
```

5. To delete the resources allocated:

```sh
chaos terminate
```


# Test It Out Manually #
0. Open a Terminal, and make sure your CRDs are compiled properly before installing them:

```sh
make generate manifests
go mod tidy
```

1. Create a Cluster using Kind (our cluster name is sample-test):

```sh
kind create cluster --name sample-test
```

2. Install Cert-Manager into the cluster:

```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.0/cert-manager.yaml
```

3. Open a new Terminal, Install the CRDs into the cluster:

```sh
make deploy
```

4. In this new Terminal, Run your controller (this will run in the foreground, so switch to the old terminal where you created the cluster if you want to leave it running):

```sh
make run
```

5. In your older Terminal, create a namespace for our issuer and certificate (the name of our namespace is chaos here):

```sh
kubectl create namespace chaos
```

6. Install Chaos Issuer using the yaml file provided in the /config/samples:

```sh
kubectl apply -f config/samples/self-signed-issuer_v1alpha1_chaosissuer.yaml
```

7. Deploy a certificate to be signed by our self-signed Chaos Issuer using the yaml file provided in the /config/samples:

```sh
kubectl apply -f config/samples/certificate_chaosissuer.yaml
```

**NOTE:** You can also run this in one step by running: `make install run`

## Modifying the API definitions ##
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Contribute ##

check out [EXTERNAL PULL REQUEST](https://github.com/kxk-4498/Venafi-test-wizard/blob/main/.github/workflows/PULL_REQUEST_TEMPLATE) on how to submit a pull request to this project.

## License  ##

Copyright 2022 CMU-SV.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
