package webauthn

import (
	"strconv"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/modulix-systems/goose-talk/internal/entity"
)

type webauthnUserAdapter struct {
	user *entity.User
}

func (u *webauthnUserAdapter) WebAuthnID() []byte {
	return []byte(strconv.Itoa(u.user.Id))
}

func (u *webauthnUserAdapter) WebAuthnName() string {
	return u.user.Email
}

func (u *webauthnUserAdapter) WebAuthnDisplayName() string {
	return u.user.Email
}

func (u *webauthnUserAdapter) WebAuthnCredentials() []webauthn.Credential {
	res := make([]webauthn.Credential, len(u.user.PasskeyCredentials))
	for _, cred := range u.user.PasskeyCredentials {
		adaptedTransports := make([]protocol.AuthenticatorTransport, len(cred.Transports))
		for _, transport := range cred.Transports {
			adaptedTransports = append(adaptedTransports, protocol.AuthenticatorTransport(transport))
		}
		res = append(res, webauthn.Credential{
			ID:        []byte(cred.ID),
			PublicKey: []byte(cred.PublicKey),
			Transport: adaptedTransports,
		})
	}
	return res
}
