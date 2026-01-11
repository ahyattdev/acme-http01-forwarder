# acme-http01-forwarder
Reverse proxy for forwarding ACME HTTP01 challenges and no other requests. This is useful for allowing HTTP01 validation challenges to reach your web server without exposing the web server itself.

You can set the target host for validation requests by setting `TARGET_HOST` to an IP, hostname, or `hostname:port`.

Docker container available via GHCR:

```shell
docker pull ghcr.io/ahyattdev/acme-http01-forwarder:<tag>
```
