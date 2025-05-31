package webauthn

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
)

type WebAuthnProvider struct {
	webAuthn *webauthn.WebAuthn
}

func New(permittedOrigins []string, displayName string, fqdn string) *WebAuthnProvider {
	wconfig := &webauthn.Config{
		RPDisplayName: displayName,
		RPID:          fqdn, // Fully Qualified Domain Name
		RPOrigins:     permittedOrigins,
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		panic(err)
	}
	return &WebAuthnProvider{webAuthn: webAuthn}
}

func (p *WebAuthnProvider) GenerateRegistrationOptions(user *entity.User) (gateways.WebAuthnRegistrationOptions, *gateways.PasskeyTmpSession, error) {
	opts, session, err := p.webAuthn.BeginRegistration(&webauthnUserAdapter{user})
	if err != nil {
		return nil, nil, err
	}
	serializedOpts, err := json.Marshal(opts)
	if err != nil {
		return nil, nil, err
	}
	adaptedCredParams := make([]gateways.PasskeyCredentialParam, len(session.CredParams))
	for _, param := range session.CredParams {
		adaptedCredParams = append(adaptedCredParams,
			gateways.PasskeyCredentialParam{Type: string(param.Type), Alg: int(param.Algorithm)},
		)
	}
	return serializedOpts, &gateways.PasskeyTmpSession{UserId: session.UserID, Challenge: session.Challenge, CredParams: adaptedCredParams}, nil
}
