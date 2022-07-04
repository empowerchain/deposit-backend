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
- A base64 encoded JSON string that is the payload: 
  - The secp256k1 public key
  - The client name (for now, needs to be 'empower-deposit-app')
  - A timestamp
- A signature of the above data

An example of the final encoded token could look like this:
```
ewogICJwYXlsb2FkIjogImV3b2dJQ0p3ZFdKTFpYa2lPaUFpTURJNU5ESXdOVE0zTkdJeU1qTTJNR05tTXprM1kyUTRNVFV3T1dGbU9HUTROalpqT1dJNE9UVXpNRFUzTkRKa1l6aGlZek14WkdVMVlUSTJabU5rT1RFMElpd0tJQ0FpWTJ4cFpXNTBJam9pWlcxd2IzZGxjaTFrWlhCdmMybDBMV0Z3Y0NJc0NpQWdJblJwYldWemRHRnRjQ0k2TVRZMU5qazBNek13TVFwOSIsCiAgInNpZ25hdHVyZSI6ICI5NDkwMjRkM2RlZDBiNDQzMTZkM2ZmYzU4N2FhZWI0MzVmOGIwZDEzYmRlODY3MzcwMGI2NzlhNTM0NDMxZmJjMjFhN2I1ZTUwOTM0ZTU1ZGEyODE2NzJlZjNiMGRjMjhmOWU3MjA0MDA2ZDM5NTZkZmU1NzJjNDFhMjdmMjE0OSIKfQ==
```

When you base64 decode it you would get something like this:
```json
{
  "payload": "ewogICJwdWJLZXkiOiAiMDI5NDIwNTM3NGIyMjM2MGNmMzk3Y2Q4MTUwOWFmOGQ4NjZjOWI4OTUzMDU3NDJkYzhiYzMxZGU1YTI2ZmNkOTE0IiwKICAiY2xpZW50IjoiZW1wb3dlci1kZXBvc2l0LWFwcCIsCiAgInRpbWVzdGFtcCI6MTY1Njk0MzMwMQp9",
  "signature": "949024d3ded0b44316d3ffc587aaeb435f8b0d13bde8673700b679a534431fbc21a7b5e50934e55da281672ef3b0dc28f9e7204006d3956dfe572c41a27f2149"
}
```

And the payload itself after being decoded could look something like this:
```json 
{
  "pubKey": "0294205374b22360cf397cd81509af8d866c9b895305742dc8bc31de5a26fcd914",
  "client":"empower-deposit-app",
  "timestamp":1656943301
}
```

### The signature

The signature is a cryptographic signature of the base64 encoded `payload` JSON string.

## Examples

To find examples of how to build up the token, take a look at `auth_test.go`
