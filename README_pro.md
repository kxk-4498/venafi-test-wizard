```sh
kubectl apply -f prometheus_setup/crd-setup
```sh
kubectl apply -f prometheus_setup/prometheus-operator-set-up
```sh
kubectl apply -f prometheus_setup/prometheus-set-up
```sh
kubectl get pods -n monitoring
