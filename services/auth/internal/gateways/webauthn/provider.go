package webauthn

import (
	"encoding/json"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
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

func (p *WebAuthnProvider) VerifyRegistrationOptions(userId int, rawCredential []byte, prevSession *gateways.PasskeyTmpSession) (*entity.PasskeyCredential, error) {
	webauthnUser := webauthnUserAdapter{user: &entity.User{ID: userId}}
	var ccr protocol.CredentialCreationResponse
	if err := json.Unmarshal(rawCredential, &ccr); err != nil {
		return nil, fmt.Errorf("%w: malformed json", gateways.ErrInvalidCredential)
	}
	parsedCredential, err := ccr.Parse()
	if err != nil {
		return nil, fmt.Errorf("%w: malformed data", gateways.ErrInvalidCredential)
	}
	adaptedCredParams := make([]protocol.CredentialParameter, len(prevSession.CredParams))
	for _, param := range prevSession.CredParams {
		adaptedCredParams = append(adaptedCredParams,
			protocol.CredentialParameter{
				Type:      protocol.CredentialType(param.Type),
				Algorithm: webauthncose.COSEAlgorithmIdentifier(param.Alg)},
		)
	}
	credential, err := p.webAuthn.CreateCredential(&webauthnUser, webauthn.SessionData{
		Challenge:      prevSession.Challenge,
		RelyingPartyID: p.webAuthn.Config.RPID,
		UserID:         prevSession.UserId,
		CredParams:     adaptedCredParams,
	}, parsedCredential)
	if err != nil {
		return nil, fmt.Errorf("%w: verification failed", gateways.ErrInvalidCredential)
	}
	adaptedTransports := make([]entity.PasskeyAuthTransport, len(credential.Transport))
	for _, transport := range credential.Transport {
		adaptedTransports = append(adaptedTransports, entity.PasskeyAuthTransport(transport))
	}
	return &entity.PasskeyCredential{
		ID:         credential.ID,
		PublicKey:  credential.PublicKey,
		UserId:     userId,
		BackedUp:   credential.Flags.BackupState,
		Transports: adaptedTransports,
	}, nil
}
