# k8s-prestop-sidecar

returns 503 to healthchecks after SIGTERM and waits until no requests arrive before exiting

- /* counts as healtcheck requests with the exception of:
- /waitz endpoint which hangs until no healthcheck requests are received
- /readyz endpoint which does NOT count as healtcheck requests

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
