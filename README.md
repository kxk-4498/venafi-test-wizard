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

Docmunet