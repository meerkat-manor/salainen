# Using Bitwarden

If you are saving your secrets in Bitwarden, then 
use this secret storage definition.

## Configuration

As the Bitwarden service is accessed via an API
the API configuration details are required.  This includes:

* URL
* User name
* Credential or API key

## Securing the Credential or API key

As Bitwarden requires credentials itself these needs
to be stored locally.  One option is to store them
in a credential file or environment variable.

The credential file or environment variable can be
accessed using **salainen**


