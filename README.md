# Venafi-test-wizard
testing tool for cert-manager in Kubernetes
# Run App in Docker
Make sure docker desktop is running in the background.<br/>
We can also run go in a small docker container: <br/>

```
cd /to/the/folder/containing/docker/file
docker build --target dev . -t go
docker run -it -v ${PWD}:/venafi go sh
go version
```
TO-DO: give documentation on how to use the issuer


# Chaos Scenario #

## Network Chaos ##

1. NETWORK DELAY

2. PACKET LOSS



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