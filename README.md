# venafi-test-wizard
Testing tool for cert-manager in Kubernetes.

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

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

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

# Chaos Scenarios #

## Network Chaos ##

1. Inject network delay inside Kubernetes Cluster 
> This test will simulate network delays between the pod Cert-manager runs on and the pod the certificate needs to renew. The user can set the network latency value , latency offset (allow the latency to fluctuate) and the duration of the test.

> A sample configuration file sets the latency to 1000 ms, offset to 200ms, making the latency to fluctuate between 800ms and 1200 ms, and duration to 10 minute is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-sample-network-delay
spec:
    action: delay
    delay:
	  latency: '1000ms' 
	  offset: '200ms'
	  duration: '10min'
```

2. Inject packet loss inside Kubernets Cluster
> This test will simulate packet loss between the pod Cert-manager runs on and the pod the certificate needs to renew. The user can set perecntage of packet loss  and the duration of the test.

> A sample configuration file sets the packet loss percentage to 50 percet,  and duration to 10 minute is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-sample-packet-loss
spec:
    action: loss
    loss:
	  percentage: '50' 
	  duration: '10min'
```

## CertificateRequest Controller Chaos ##
A normal functining CertificateRequest Controller will watch for CertificateRequest resources and attempt to sign their attached certificate signing requests (CSR). 

In the Chaos Issuer, the CertificateRequest Controller will function in a abnormal way and bring chaos into the certificate issuing process

1. Issue the CertificateRequest using a Issuer that does not belong to the request's group.

> This test will disable the code that checks whether the issuer's group matches the request's group. The issuer will accept all incomming requests without checking its group. The user can manually set the issuer's group to be anything different from the request's group.

```
	// Ignore CertificateRequest if issuerRef doesn't match our group
	if certificateRequest.Spec.IssuerRef.Group != sampleissuerapi.GroupVersion.Group {
		log.Info("Foreign group. Ignoring.", "group", certificateRequest.Spec.IssuerRef.Group)
		return ctrl.Result{}, nil
	}
```

2. Set `Ready = Signed` when CertificateRequest has been denied

>This test disabled the below code blocks and forces denied CertificateRequests to be signed.
```
	// If CertificateRequest has been denied, mark the CertificateRequest as
	// Ready=Denied and set FailureTime if not already.
	if cmutil.CertificateRequestIsDenied(&certificateRequest) {
		log.Info("CertificateRequest has been denied yet. Marking as failed.")

		if certificateRequest.Status.FailureTime == nil {
			nowTime := metav1.NewTime(r.Clock.Now())
			certificateRequest.Status.FailureTime = &nowTime
		}

		message := "The CertificateRequest was denied by an approval controller"
		setReadyCondition(cmmeta.ConditionFalse, cmapi.CertificateRequestReasonDenied, message)
		return ctrl.Result{}, nil
	}

	if r.CheckApprovedCondition {
		// If CertificateRequest has not been approved, exit early.
		if !cmutil.CertificateRequestIsApproved(&certificateRequest) {
			log.Info("CertificateRequest has not been approved yet. Ignoring.")
			return ctrl.Result{}, nil
		}
	}
```
> The code below forces the request to be signed
```
	certificateRequest.Status.Certificate = signed

	setReadyCondition(cmmeta.ConditionTrue, cmapi.CertificateRequestReasonIssued, "Signed")
	return ctrl.Result{}, nil
```

3. Force the controller to sleep periodically

> The controller will stop responding incomming and outgoing requests during sleep time. User can set the interval between sleeps and the duration of each sleep. 

## License

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

