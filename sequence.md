# Sequence Diagram #
## Shows all possible commands which can be given by a developer to our shell script. ##
## About choas scenarios: ##
How chaos scenarios will work is that the developer will provide the parameters inside the issuer.yaml file and once it reaches our controller manager, it takes action based on the inputs written in the file. For example, if the developer wants to make the issuer sleep for X amount of time before signing the certificate, then it can pass that parameter as: 
```sh
apiVersion: self-signed-issuer.chaos.ch/v1alpha1
kind: ChaosIssuer
metadata:
  name: chaosissuer-scenario3
spec:
  selfSigned: {}
  Scenario3:
    duration: 10
```
Now everytime, the issuer signs the certificate, it'll sleep for 10 seconds before signing it. the delay will be logged as well for getting metrics.

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
    Chaos Issuer CLI->>Kubernetes API:make run
    Developer->>Chaos Issuer CLI:chaos deploy issuer
    Chaos Issuer CLI->>Kubernetes API:kubectl apply chaos issuer
    Kubernetes API->>Chaos Controller Manager:create chaos issuer
    Chaos Controller Manager->>Kubernetes API:choas issuer ready
    Note over Kubernetes API, Chaos Controller Manager:Chaos Issuer is deployed with the chaos scenarios mentioned in yaml file of the Issuer
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
```

