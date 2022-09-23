# Venafi-test-wizard
testing tool for cert-manager in Kubernetes
# Run App in Docker
We can also run go in a small docker container: <br/>

```
docker build --target dev . -t go
docker run -it -v ${PWD}:/work go sh
go version
```