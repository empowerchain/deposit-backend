# Auth

We're using a bit of a homegrown auth solution based on secp256k1 key-pairs and signtures.

The whole reason we are doing this is to 'emulate' a blockchain backend and more importantly
let the frontend only authenticate with the same keypair as they use for the Cosmos ecosystem.

## Transport

The transport mechanism is using the same as the traditional JWT header with a 'Bearer' keyword.

So the token we send with every request to authenticate the user will be on:
- Header name: Authorization
- Header content: 'Bearer TOKEN'

## Token structure

The token itself is a base64 encoded JSON structure that includes:
- The secp256k1 public key
- The client name (for now, needs to be 'empower-deposit-app')
- A timestamp
- A signature of the above data

An example of the final encoded token could look like this:
```
ewogICJwYXlsb2FkIjogeyJwdWJLZXkiOiIwMzlmODgyNTdmZDdmNjNiNTM1MTY4YWE1YjU1NjEzMzE3NDE1OTcwNWQxMDJmM2M5M2IxNTdiYzhjYTUzOWI2ZTciLCJjbGllbnQiOiJlbXBvd2VyLWRlcG9zaXQtYXBwIiwidGltZXN0YW1wIjoxNjU2OTMxNDkzfSwKICAic2lnbmF0dXJlIjogIjgyNDYzZjFhOTY0NzZlNGY3OGJjOWYxZDAxNTUyZDFiMmZmYjQ1OGU1ZTE3OWZjNjgxZGUxOTJjODY5YmQ2NzIyZjViNzVkNTI2OGYzZjU2MWY1YjgxYmY0NTdkOTk1Y2MyZTZjMTFkMTE3ODg0YmFhYzhiYzg2ZGI1ZTRmYzMyIgp9
```

When you base64 decode it you would get something like this:
```json
{
  "payload": {
    "pubKey":"039f88257fd7f63b535168aa5b556133174159705d102f3c93b157bc8ca539b6e7",
    "client":"empower-deposit-app",
    "timestamp":1656931493
  },
  "signature": "82463f1a96476e4f78bc9f1d01552d1b2ffb458e5e179fc681de192c869bd6722f5b75d5268f3f561f5b81bf457d995cc2e6c11d117884baac8bc86db5e4fc32"
}
```

### The signature

The signature is a cryptographic signature of the `payload`.

It is very important that the payload signature is made on a string of the json format without any line breaks or spaces, like this:
```
{"pubKey":"039f88257fd7f63b535168aa5b556133174159705d102f3c93b157bc8ca539b6e7","client":"empower-deposit-app","timestamp":1656931493}
```

## Examples

To find examples of how to build up the token, take a look at `auth_test.go`
