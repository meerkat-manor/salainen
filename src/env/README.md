# Environment storage


## Windows environment

If you are using OS environment variables and
on Windows, then from a command line you need to
quote the key.

Your request for a secret value stored in the 
Windows environment variable using the command line
is:

```
salainen '$env:MY_SECRET'
```

as an example for enviroment variable ```MY_SECRET``

## Linux


## Alternative

An alternative form exists for environment variables
by surrounding the variable name with `${` and `}`

```
salainen '${MY_SECRET}`
```

