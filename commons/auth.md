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
ewogICJwYXlsb2FkIjogImV3b2dJQ0p3ZFdKTFpYa2lPaUFpTUROak9UWTBPVFZoTWpSbFpUWXpOalZsWlRFelpUTXhOMlUxWmpWak1UTTNaRGczTW1JeFl6ZzVaR1poWkRCa05HSmtZekUzTnpnME56VmlOV0ZpTjJKa0lpd0tJQ0FpWTJ4cFpXNTBJam9pWlcxd2IzZGxjaTFrWlhCdmMybDBMV0Z3Y0NJc0NpQWdJblJwYldWemRHRnRjQ0k2TVRZMU5qa3pPVGt4TVFwOSIsCiAgInNpZ25hdHVyZSI6ICI1MDQzNjE0ZWVhMTM0NTc2MGExNzk5NGRhYTBmMDcyZWNlNGE1ZGMxOTQ1MzQwMDk3MTZjYzczM2E4ZTNhZWEzNmE3MGU3ZmU2NDU5NGIzYmQ3MmRhYTIzODJlNjVjYWI1MmY5NWNmYWE1ZWRmY2UzZGEyODAwNjBmY2EzNTZjNyIKfQ==
```

When you base64 decode it you would get something like this:
```json
{
  "payload": "ewogICJwdWJLZXkiOiAiMDNjOTY0OTVhMjRlZTYzNjVlZTEzZTMxN2U1ZjVjMTM3ZDg3MmIxYzg5ZGZhZDBkNGJkYzE3Nzg0NzViNWFiN2JkIiwKICAiY2xpZW50IjoiZW1wb3dlci1kZXBvc2l0LWFwcCIsCiAgInRpbWVzdGFtcCI6MTY1NjkzOTkxMQp9",
  "signature": "5043614eea1345760a17994daa0f072ece4a5dc194534009716cc733a8e3aea36a70e7fe64594b3bd72daa2382e65cab52f95cfaa5edfce3da280060fca356c7"
}
```

And the payload itself after being decoded could look something like this:
```json 
{
  "pubKey": "03c96495a24ee6365ee13e317e5f5c137d872b1c89dfad0d4bdc1778475b5ab7bd",
  "client":"empower-deposit-app",
  "timestamp":1656939911
}
```

### The signature

The signature is a cryptographic signature of the base64 encoded `payload` JSON string.

## Examples

To find examples of how to build up the token, take a look at `auth_test.go`
