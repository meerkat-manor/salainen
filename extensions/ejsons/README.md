# EJSON Git storage file

The ejson storage method for secrets will define a special
JSON formatted file that can be stored in GIT because the values
are encrypted, but the attribute names are not.

See [https://github.com/Shopify/ejson](https://github.com/Shopify/ejson) for
more details on the formatting.

## Summary of ejson format

Secrets are collected in a JSON file, in which all the string values are encrypted.
Public keys are embedded in the file, and the decrypter looks up the corresponding
private key from its local filesystem.

_Note_: YAML formatted files are not supported.

## Configuration

_Warning_: If the private key is supplied in the configuration and not through
the Key directory, then value needs to be secured by another type of
**salainen** provider to ensure secrecy and not expose the key in configuration.  You could 
use providers such as "keyring:" and "wincred:" and of course "plain:" . A "plain:"
provider should only be used for testing **salainen** not for other
situations.

## Usage

Store and version the file in Git or other locations as required, but ensure
secrets are encrypted before storage, when you require a value, call the 
**Get** function passing the file (as key) and private key in the custom
configuration for **ejson**.

Only one element can be decrypted and this is passed as the **ElementName**
in the custom configuration.

### Scenario with configuration file

1. Create a configuration file and include an element name 
   _public_key at the root
2. This configuration file can be saved in Git with the project.
   This could a default or template configuration file.
3. Include all the other element names with leading "_" symbol
   on the name.  The prefix "_" indicates to the ejson processing
   that the value is not to be encrypted / decrypted.
4. Include the secret, say name, "password" element name.  This value
   is then encrypted and committed to Git
5. Whenever you need to use the secret within your Golang
   program or script, call the **salainen** CLI/Get function
   passing in the JSON configuration file name and the custom config
   such as directory and private key
6. The Get function will return the secret in the clear.

_Note_: It is important that you encrypt the JSON configuration file
before committing to Git.  This can be part of your pipeline process.
You can use the **salainen** CLI Put function to achieve this.

