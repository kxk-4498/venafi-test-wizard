# Sequence Diagram #
## Shows all possible commands which can be given by a developer to our shell script. ##
## About choas scenarios: ##
How chaos scenarios will work is that the developer will provide the parameters inside the issuer.yaml file and once it reaches our controller manager, it takes action based on the inputs written in the file. For example, if the developer wants to make the issuer sleep for X amount of time before signing the certificate, then it can pass that parameter as: 
```sh
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario
spec:
  selfSigned: {}
  Scenarios:
    sleepDuration: "10"
```
Now everytime, the issuer signs the certificate, it'll sleep for 10 seconds before signing it. the delay will be logged as well for getting metrics.
**NOTE:** chaos show metrics is under progress and not available in PROD yet.

```mermaid
sequenceDiagram
    actor Developer
    autonumber
    participant Chaos Issuer CLI
    participant Kubernetes API
    participant Cert Manager
    participant Chaos Controller Manager
    Developer->>Chaos Issuer CLI:chaos setup
    Chaos Issuer CLI->>Kubernetes API:kind create cluster
    Chaos Issuer CLI->>Kubernetes API:kubectl apply cert-manager
    Kubernetes API->>Cert Manager:creating cert manager resources
    Cert Manager->>Kubernetes API:cert manager resources created and ready
    Chaos Issuer CLI->>Kubernetes API:make deploy
    Kubernetes API->>Chaos Controller Manager:created custom resources
    Chaos Controller Manager->>Kubernetes API:custom resources created
    Chaos Issuer CLI->>Kubernetes API:make run
    Kubernetes API->>Chaos Controller Manager:Run controller manager
    Chaos Controller Manager->>Kubernetes API:chaos controller manager live and running
    Kubernetes API->>Chaos Issuer CLI:all resources created, live and running
    Chaos Issuer CLI-->>Developer:return
    Developer->>Chaos Issuer CLI:chaos deploy issuer
    Chaos Issuer CLI->>Kubernetes API:kubectl apply chaos issuer
    Kubernetes API->>Chaos Controller Manager:create chaos issuer
    Chaos Controller Manager->>Kubernetes API:choas issuer ready
    Note over Kubernetes API, Chaos Controller Manager:Chaos Issuer is deployed with the chaos scenarios mentioned in yaml file of the Issuer
    Kubernetes API->>Chaos Issuer CLI:choas issuer ready
    Developer->>Chaos Issuer CLI:chaos deploy cert
    Chaos Issuer CLI->>Kubernetes API:kubectl apply certificate
    Kubernetes API->>Cert Manager:creating certificate
    Cert Manager-->>Cert Manager:creating temp private key
    Cert Manager-->>Cert Manager:creating certificate request using key and certificate
    Cert Manager->>Chaos Controller Manager:sending certificate request
    Chaos Controller Manager->>Cert Manager:certificate request signed
    Cert Manager->>Kubernetes API:Certificate ready
    Developer->>Chaos Issuer CLI:chaos show
    Chaos Issuer CLI->>Kubernetes API:kubectl get cert
    Kubernetes API->>Chaos Issuer CLI:list of certs
    Chaos Issuer CLI->>Developer:list of certs
    Developer-->>Chaos Issuer CLI:chaos show metrics
    Chaos Issuer CLI-->>Kubernetes API:kubectl logs cert-manager
    Kubernetes API-->>Cert Manager:get cert-manager logs
    Cert Manager-->>Chaos Issuer CLI: show cert-manager logs
    Chaos Issuer CLI-->>Kubernetes API:kubectl logs controller-manager
    Kubernetes API-->>Chaos Controller Manager:get controller manager logs
    Chaos Controller Manager-->>Chaos Issuer CLI: show chaos controller-manager logs
    Chaos Issuer CLI-->>Developer: Logs
    Developer->>Chaos Issuer CLI:chaos terminate
    Chaos Issuer CLI->>Kubernetes API: kubectl delete cluster
    Kubernetes API->>Chaos Issuer CLI: cluster deleted
    Chaos Issuer CLI->>Developer: all resources deleted
```

