# Prompt

This type of secret storage is not actually a storage but a user prompt for the 
secret.

**Warning:** This secret storage type is not suitable for headless scenarios.

## Definition

Following the secret storage definition convention you can have:


```
prompt:secret_password
```

which provides a default secret of ``secret_password`` if no value manually
input. As the default value is fixed this method has little value.

**Warning:** The above definition is not very secure and is equivalent to the
[plain](../plain/) secret storage type

_OR_


```
prompt:
```

which requires the manual input of a secret. That is its mandatory to supply a value.

You can cancel the prompt with a Ctrl+C.

## Rational

This secret storage type can be useful when running a batch or service locally in
non daemon mode and can be prompted for password.  This is likely to be the case
for local development and debugging of code.

## Configuration and usage

The prompt secret type can be configured to require a certain complexity before
teh value is accepted.  This can include:

1. minimum and maximum length
2. acceptable character set

The validation is defined by a Golang regex expression.

The put or save action for ``prompt`` is not available as there is no
secret storage location just a definition.
