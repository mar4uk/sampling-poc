# Sampling POC

## How to run
```shell
docker-compose up
```

Queries to check it works:
Check that sampling affects number of traces in tempo (should be around 25)
```shell
rate(tempo_ingester_traces_created_total{job="tempo"}[5m])
```

Check that sampling doesn't affect RED metrics created by spanmetrics (should be around 50)
```shell
rate(calls{job="otel-collector",service_name="hello-world-app"}[5m])
```