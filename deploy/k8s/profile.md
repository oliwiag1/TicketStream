# Kubernetes Deployment Profile

This profile targets a baseline production cluster with separate namespaces per environment.

## Components

- `ticketstream-api` deployment + service
- `ticketstream-worker` deployment
- Ingress for frontend/API
- External managed PostgreSQL/Redis/RabbitMQ recommended

## Apply

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/api.yaml
kubectl apply -f deploy/k8s/worker.yaml
kubectl apply -f deploy/k8s/ingress.yaml
```

## Notes

- Secrets should be injected via external secret manager.
- Readiness probes use `/health/ready`.
- Liveness probes use `/health/live`.
- HPA can scale API on CPU or custom request metrics.
