package sign

import (
	"context"
)

// RPCAuthOIDC is used to perform the RPC login using OIDC instead of a fixed token.
type RPCAuthOIDC struct {
	Path  string // path defaults to "jwt" for vault
	Role  string // role is required for jwt logins
	Token string // token is a jwt with vault
}

// RPCAuth provides credentials for RPC calls, empty fields are ignored.
type RPCAuth struct {
	Address string // address is the remote server address, e.g. https://vault:8200
	Path    string // path for the RPC, in vault this is the transit path which default to "transit"
	Token   string // token used for RPC, in vault this is the VAULT_TOKEN value
	OIDC    RPCAuthOIDC
}

// RPCOption specifies options to be used when performing RPC.
type RPCOption interface {
	ApplyContext(*context.Context)
	ApplyRemoteVerification(*bool)
	ApplyRPCAuthOpts(opts *RPCAuth) // was options.RPCAuth
	ApplyKeyVersion(keyVersion *string)
}

// PublicKeyOption specifies options to be used when obtaining a public key.
type PublicKeyOption interface {
	RPCOption
}
