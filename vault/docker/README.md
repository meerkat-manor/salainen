

Instructions followed from https://hub.docker.com/r/hashicorp/vault


```
docker run --cap-add=IPC_LOCK -d --name=dev-vault -p 8200:8200 hashicorp/vault
```

docker run --cap-add=IPC_LOCK -e 'VAULT_LOCAL_CONFIG={"storage": {"file": {"path": "/vault/file"}}, "listener": [{"tcp": { "address": "0.0.0.0:8200", "tls_disable": true}}], "default_lease_ttl": "168h", "max_lease_ttl": "720h", "ui": true}' -p 8200:8200 hashicorp/vault server

