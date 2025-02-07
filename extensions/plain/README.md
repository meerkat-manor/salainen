# Plain 

This is plain password in the path and is not the path key to
the storage.  This type of secret storage is insecure as the
value is fixed to where the secret type is stored.

For example if the secret type is stored in a configuration file
then the secrecy of the value is only as good as the security
of the configuration file.

**Warning:** Do not use this secret storage type in production
environments.

## Rational

This secret storage type is provided so that the ``salainen`` fetch
can be used consistently such as for temporary or local secrets
during development.

## Configuration and usage

The put or save action for ``plain`` is not available as there is no
secret storage location just a definition.
