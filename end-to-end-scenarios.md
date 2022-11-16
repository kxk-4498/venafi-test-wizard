# Sequence Diagram for end to end scenarios #
## What has been done so far: ##

The team developed a configurable Chaos Issuer where chaotic scenarios and their parameters can be given inside the yaml file of the issuer or given as an command in the chaos issuer CLI. The Chaos Issuer will then sign the certificate based on the chaotic scenario it is deployed with.

## Question: ##
Jing - "I see a clear separation from the target which is the system/cluster that’s been tested and the chaos test tool which is the software a QA or a Reliability engineer uses to test the robustness and resilience of the target. 

What do you think you are building? the target or the tool or both or neither? Who will be using your software? a developer, a reliability engineer or some others?".

## Our solution and approach to the question: ##

The team saw that the end-to-end component was incomplete and we are proposing the below solution to resolve this.

End Scenario script - End Scenario script contains commands for every chaotic scenario to reach its end state. It gets triggered by Chaos Issuer getting deployed with chaotic scenario flags. Only After some conditions are met such as waiting for the Chaos Issuer to sign a certificate, the script launches the set of commands to cause chaos and reach an end state and simultaneously pull logs and make a report of the whole scenario which can be shown to developer later on.

## An example of end-to-end scenarios: ##
1. Launch the base scenario working with our Chaos Issuer with all the chaotic scenario flags turned off and issue a certificate which can be used by the application. Say for example, a certificate with duration of 1 hour and a renewal time of 30 mins.
2. Now, the developer wants to launch a chaos scenario and deploys an issuer with the chaotic scenario flag set. For example, the issuer sleeps for X amount of time. **NOTE:** The end scenario script will get triggered and will be get ready to cause the end scenario to take place for issuer sleeping.
3. Now when the certificate attempts to renew, it hits the Chaos Issuer with the specfic chaos scenario flag set. If the chaos scenario selected was, say an issuer that delays the certificate signing process by 1 hour, the renewal attempt is made with 30 minutes left to expiry. If the Chaos Issuer sleeps for 1 hour, this would essentially mean the certificate signing process will not get finished before the certificate expires. **NOTE:** In order to speed up this part, the end scenario script will force renewal of the certificate.
4. This causes the application to no longer work in a secure manner as the certificate is now expired.
5. The end scenario script perodically parses the various logs and creates a final report depicting the various scenarios that happened and the chain of events which led to failure.
6. This report can be viewed by the developer to understand the whole scenario and the end result.

## Sequence Diagram from step 2 of the end to end scenario given above: ##

```mermaid
sequenceDiagram
    actor Developer
    autonumber
    participant Chaos Issuer CLI
    participant E as End Scenario Script
    participant Kubernetes API
    participant Cert Manager
    participant Chaos Controller Manager
    participant WebApplication
    Cert Manager->>Cert Manager:Signed certificate exists in secret
    Note right of Cert Manager: certificate valid for 10 mins
    WebApplication->>Cert Manager:Use certificate to function successfully
    Developer->>Chaos Issuer CLI:chaos deploy issuer scenario 3
    Chaos Issuer CLI->>Kubernetes API:kubectl apply chaos issuer
    Kubernetes API->>Chaos Controller Manager:create chaos issuer
    Chaos Controller Manager->>Kubernetes API:choas issuer ready
    Note over Kubernetes API, Chaos Controller Manager:Chaos Issuer is deployed with the chaos scenario flag in yaml file of Issuer or using Choas Issuer CLI command
    Kubernetes API->>Chaos Issuer CLI:choas issuer ready
    Note over Chaos Issuer CLI,Kubernetes API:End Script Scenario Triggered as conditions of existing signed certificate and issuer deployed with sleeping flag are met.
    E->>Kubernetes API:cmctl renew certificate
    Kubernetes API->>Cert Manager:renewng certificate
    Cert Manager-->>Cert Manager:reusing temp private key
    Cert Manager-->>Cert Manager:creating certificate request using key and certificate
    Cert Manager->>Chaos Controller Manager:sending certificate request 
    Chaos Controller Manager->>Chaos Controller Manager:sleep for 10 mins
    Cert Manager->>Cert Manager: certificate expired
    Note over WebApplication, Cert Manager: Unable to use expired certificate, application unable to use HTTPS.
    Chaos Controller Manager->>E:Fetching Logs
    Cert Manager->>E:Fetching Logs
    WebApplication->>E:Fetching Logs
    E->>E:creating report
    E->>Chaos Issuer CLI:Report Sent
    Developer->>Chaos Issuer CLI: chaos show report scenario 3
    Chaos Issuer CLI->>Developer: showing report
    Chaos Controller Manager->>Cert Manager:certificate request signed after sleeping for 10 mins
    Cert Manager->>Kubernetes API:Certificate renewed
    WebApplication->>Cert Manager:Using renewed certificate to function successfully
```