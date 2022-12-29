# k8s-prestop-sidecar

why: AWS NLB ignores Kubernetes Endpoints and is slow to react.

Set this sidecar to be the health check port and it will returns 503 to healthchecks after SIGTERM and waits until no requests arrive before exiting after COOLDOWN duration

## Routes

- /* counts towards healtcheck requests with the exception of:
- /waitz endpoint which hangs until no healthcheck requests are received and COOLDOWN is slept
- /readyz endpoint which does NOT count towards healtcheck requests and returns 503 after SIGTERM
- /healthz endpoint which always returns 200 and does NOT count towards healthcheck requests

## Example with ingress-nginx-controller

Add preStop hook that calls this sidecar

```yaml
  lifecycle:
    preStop:
      exec:
        command:
          - "/usr/bin/curl"
          - "localhost:8080/waitz"
```

And add sidecar

```yaml
  extraContainers:
    - name: k8s-prestop-sidecar
      image: ghcr.io/matti/k8s-prestop-sidecar:742dbb80dc68734547db97ad318df705f52bc7bd
      env:
        - name: "LOG"
          value: "yes"
        - name: "COOLDOWN"
          value: "120s"
      readinessProbe:
        httpGet:
          path: /readyz
          port: 8080
        initialDelaySeconds: 5
        timeoutSeconds: 1
        periodSeconds: 3
        successThreshold: 1
        failureThreshold: 1
```

What happens:

- ingress-nginx-controller executes `preStop` which hangs and ingress-nginx keeps serving requests until they stop
- k8s-prestop-sidecar will receive SIGTERM and return 503 for /readyz which will start removing the endpoint and loadbalancer targets
- after k8s-prestop-sidecar receives no traffic, it will still wait COOLDOWN until releases /waitz hang and exits.

## log sample

```console
$ INTERVAL=6s LOG=yes go run main.go
2022/12/29 11:54:48 k8s-prestop-sidecar started
2022/12/29 11:54:49 hit / from 127.0.0.1:56946 status 200 current rate 1 requests in 6s
2022/12/29 11:54:52 hit / from 127.0.0.1:56957 status 200 current rate 2 requests in 6s
2022/12/29 11:54:54 received signal term
2022/12/29 11:54:54 shutdown current rate 2 requests in 6s
2022/12/29 11:54:55 hit / from 127.0.0.1:56966 status 503 current rate 2 requests in 6s
2022/12/29 11:54:55 shutdown current rate 2 requests in 6s
2022/12/29 11:54:56 shutdown current rate 2 requests in 6s
2022/12/29 11:54:57 shutdown current rate 2 requests in 6s
2022/12/29 11:54:58 shutdown current rate 1 requests in 6s
2022/12/29 11:54:59 shutdown current rate 1 requests in 6s
2022/12/29 11:55:00 shutdown current rate 1 requests in 6s
2022/12/29 11:55:01 shutdown current rate 0 requests in 6s
2022/12/29 11:55:01 bye
```
