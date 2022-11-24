```
kubectl apply -f prometheus_setup/crd-setup
```
kubectl apply -f prometheus_setup/prometheus-operator-set-up
```
kubectl apply -f prometheus_setup/prometheus-set-up
```
kubectl get pods -n monitoring
