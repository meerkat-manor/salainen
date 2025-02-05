# Salainen command

The **salainen** comman line (CLI) function is provided to enable 
a simple functon to set and get secrets (passwords, cerdentials).

There are a number of alternatives available especially if you are
after a UI.

The **salainen** is primarily intended to be used in Golang projects as 
a package.

## Objective

The idea of the package was drive out of a need to use secrets for 
various assets, such as a database, web service or email service.

A common approach is to use environment variables, but the values should
not be captured in the source or IDE files that are saved to Git 
repositories. This presents a challege as you cannot use scripts for
local execution or testing very easily.

You could use something like Vault for all your projects, but this could be 
over engineering for many small projects.  You could use the native 
cloud key management / secrets solution, but again this might 
not suit for local development.

The secrets also need to reognise that there are different environments
such as local development, build pipelines, testing, staging and finally
production.

Your developers may also on different OSs such as Lnux, Mac or Windows. 

Each of these scenarious can have different secret storage capabilities and
user needs.

## Use case

### Local development scripts

###  Scripting

As the **salainen** program can exceute natively on a many platforms, you can use
it in scripts such as ``bash`` or ``powershell`` to fetch secrets and then use 
secret value in accessing secured assets.

Commonly you might save secrets in environment variables, and with **salainen**
you can continue to do so, but you now have the option of making them more
dynamic or slecting alternative, more secure methods.

You can pipe the secret from the command like so:

```bash
salainen --config ~/.secrets/salainen.json keyring:db_password > 
```

```powershell
salainen --config ~\.secrets\salainen.json wincred:db_password > 
```

### Docker environments

### Build pipelines

### Test execution

### Production execution

The **salainen** CLI is not intended to be used in the daily, headless execution of
code whether it be a CRON job or a web asset.

For these situations you should use the Golang **salainen** package within your 
Golang application.

You can use the **salainen** CLI to set the secret values that are then consumed
by you program.
