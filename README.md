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

# Chaos Scenarios #

## Network Chaos ##

1. Inject network delay inside Kubernetes Cluster 
> This test will simulate network delays between the pod Cert-manager runs on and the pod the certificate needs to renew. The user can set the network latency value , latency offset (allow the latency to fluctuate) and the duration of the test.

> A sample configuration file sets the network delay to 1000ms is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario2
spec:
  selfSigned: {}
  Scenarios:
    sleepDuration: "0"
    Scenario1: "False"
    Scenario2: "False"
	networkDelay: "1000" 
```

2. Inject packet loss inside Kubernets Cluster
> This test will simulate packet loss between the pod Cert-manager runs on and the pod the certificate needs to renew. The user can set perecntage of packet loss  and the duration of the test.

> A sample configuration file sets the packet loss percentage to 50 percet is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario2
spec:
  selfSigned: {}
  Scenarios:
    sleepDuration: "0"
    Scenario1: "False"
    Scenario2: "False"
	networkDelay: "0" 
	packetLoss: "50"
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
> A sample configuration file enables this test is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario2
spec:
  selfSigned: {}
  Scenarios:
    sleepDuration: "0"
    Scenario1: "True"
    Scenario2: "False"
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
> A sample configuration file enables this test is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario2
spec:
  selfSigned: {}
  Scenarios:
    sleepDuration: "0"
    Scenario1: "False"
    Scenario2: "True"
```

3. Force the controller to sleep periodically

> The controller will stop responding incomming and outgoing requests during sleep time. User can set the interval between sleeps and the duration of each sleep. 

> A sample configuration file sets sleep duration to 10 second is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario3
spec:
  selfSigned: {}
  Scenarios:
    sleepDuration: "10"
```
> A sample configuration file set sleep time to 20 secondes test is shown below
```
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario2
spec:
  selfSigned: {}
  Scenarios:
    sleepDuration: "20"
    Scenario1: "False"
    Scenario2: "True"
```
## Code Walkthrough  ##
Our test wizzard is build on top of Cert-manager's sample external issuer https://github.com/cert-manager/sample-external-issuer.git

We added several CRDs into Kubernetes to support our Test Wizzard, the code can be found in `Venafi-test-wizard/api/v1alpha1/chaosissuer_types.go`. The code below shows how our CRDs are defined. 

A new struct called `ChaosScenarios` is defined to hold our chaos scenarios. `ChaosScenarios` contains three variables: `SleepDuration`, `Scenario1` and `Scenario2`. `SleepDuration` takes an integer while `Scenario1` and `Scenario2` take a boolean value. `ChaosScenarios` is then passed into `ChaosIssuerSpec` to support our yaml configurations.
```
// Code from ./api/v1alpha1/chaosissuer_types.go

type ChaosIssuerSpec struct {

	SelfSigned *SelfSignedIssuer `json:"selfSigned,omitempty"`
	Scenarios  *ChaosScenarios   `json:"Scenarios"`
}

type ChaosScenarios struct {
	SleepDuration string `json:"sleepDuration"`
	Scenario1     string `json:"Scenario1"` //Scenario 1: when issuer doesn't belong to request group
	Scenario2     string `json:"Scenario2"` //Scenario 2: Set Ready = Signed when CertificateRequest has been denied
}
```

The code for our chaos scenarois can be found in `Venafi-test-wizard/controllers/certificaterequest_controller.go`.

The code shown below fetches the values inside the yaml file passed in and stores them in an object called `chaosIssuer`.

```
func (r *CertificateRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := r.Log.WithValues("certificaterequests", req.NamespacedName)

	// Fetch the CertificateRequest resource being reconciled.
	// Just ignore the request if the certificate request has been deleted.
	cr := cmapi.CertificateRequest{}
	if err := r.Client.Get(ctx, req.NamespacedName, &cr); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to retrieve CertificateRequest resource")
		return ctrl.Result{}, err
	}

	chaosIssuer := api.ChaosIssuer{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: cr.Spec.IssuerRef.Name}, &chaosIssuer); err != nil {
		err := r.setStatus(ctx, log, &cr, cmmeta.ConditionFalse, cmapi.CertificateRequestReasonPending, "Failed to retrieve chaosIssuer %s/%s: %v", req.Namespace, cr.Spec.IssuerRef.Name, err)
		return ctrl.Result{}, err
	}
```

This block of codes extracts the CertificateRequest Controller Chaos values from `chaosIssuer` and stores them in two local variables called `csr1` and `csr2`.

Scenario 1: when issuer doesn't belong to request group.

Scenario 2: Set Ready = Signed when CertificateRequest has been denied
```

	csr1, err = strconv.ParseBool(chaosIssuer.Spec.Scenarios.Scenario1)
	if err != nil {
		log.Error(err, "failure")
	}

	csr2, err = strconv.ParseBool(chaosIssuer.Spec.Scenarios.Scenario2)
	if err != nil {
		log.Error(err, "failure")
	}

```
The code below checks the whether `csr1` is set to True. If `csr1` is set to True, the code will omit the request group and issuer group check and issue the certificate anyway.

```
	// Check the CertificateRequest's issuerRef and if it does not match the api
	// group name, log a message at a debug level and stop processing.
	log.V(4).Info("Scenario1 check if value csr1 is True. csr1:%t", csr1)
	if !csr1 {
		if cr.Spec.IssuerRef.Group != "" && cr.Spec.IssuerRef.Group != api.GroupVersion.Group {
			log.V(4).Info("resource does not specify an issuerRef group name that we are responsible for", "group", cr.Spec.IssuerRef.Group)
			return ctrl.Result{}, nil
		}
	}
```
The code below checks the whether `csr2` is set to True. If `csr2` is set to True, the code will always set Ready = Signed when CertificateRequest has been denied.
```
	//requestShouldBeProcessed is function given below to check for different conditions of Certificate Request
	log.V(4).Info("Scenario2 check if value csr1 is True. csr2:%t", csr2)
	if !csr2 {
		shouldProcess, err := r.requestShouldBeProcessed(ctx, log, &cr)
		if err != nil || !shouldProcess {
			return ctrl.Result{}, err
		}
	}
```

The code below fetches the sleep time from `chaosIssuer` and cast the value into a integer called `globalSleepDuration`. Then the code will manually sleep for x secondes based on `globalSleepDuration`. 
```
	globalSleepDuration, err = strconv.Atoi(chaosIssuer.Spec.Scenarios.SleepDuration)
	if err != nil {
		log.Error(err, "failure")
	}

	if globalSleepDuration != 0 {
		log.V(4).Info("default values of the chaos sleep scenario with error %ds: %s", globalSleepDuration, err)
		time.Sleep(time.Duration(globalSleepDuration) * time.Second)
		//time.Sleep(time.Duration(globalSleepDuration) * time.Second)
		//return ctrl.Result{}, nil
	}
```


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
