# EJSON Git storage file

The ejson storage method for secrets will define an special
JSON formatted file that can be stored in GIT because the values
are encrypted, but the attribute names are not.

See [https://github.com/Shopify/ejson](https://github.com/Shopify/ejson) for
more details on the formatting.

## Summary of ejson format

Secrets are collected in a JSON file, in which all the string values are encrypted.
Public keys are embedded in the file, and the decrypter looks up the corresponding
private key from its local filesystem.

