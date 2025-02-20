

Instructions followed from https://hub.docker.com/r/hashicorp/vault

https://www.virtualizationhowto.com/2025/01/hashicorp-vault-docker-install-steps-kubernetes-not-required/

docker exec -it vault sh -c 'VAULT_ADDR="http://127.0.0.1:8200" vault operator init'

Access the UI

http://127.0.0.1:8200/ui/vault/dashboard


$ docker run --cap-add=IPC_LOCK -e 'VAULT_LOCAL_CONFIG={"storage": {"file": {"path": "/vault/file"}}, "listener": [{"tcp": { "address": "0.0.0.0:8200", "tls_disable": true}}], "default_lease_ttl": "168h", "max_lease_ttl": "720h", "ui": true}' -e 'VAULT_DEV_ROOT_TOKEN_ID=salainen-root' -p 8200:8200 hashicorp/vault server -dev


docker run --cap-add=IPC_LOCK -e 'VAULT_DEV_ROOT_TOKEN_ID=myroot' -e 'VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200' -e 'VAULT_ADDR=http://0.0.0.0:8200' hashicorp/vault
