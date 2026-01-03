package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/internal/dtos"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
)

func (s *AuthService) SignUp(
	ctx context.Context,
	dto *dtos.SignUpRequest,
) (*dtos.SignUpResponse, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrOtpIsNotValid
		}
		return nil, err
	}

	err = s.securityProvider.ComparePasswords(otp.Code, dto.ConfirmationCode)
	if err != nil {
		s.log.Error("AuthService.SignUp - invalid otp", err, "err")
		return nil, ErrOtpIsNotValid
	}

	hashedPassword, err := s.securityProvider.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}
	user, err := s.usersRepo.Insert(
		ctx,
		&entity.User{
			FirstName:  dto.FirstName,
			LastName:   dto.LastName,
			Email:      dto.Email,
			Password:   hashedPassword,
			PrivateKey: s.securityProvider.GeneratePrivateKey(),
		},
	)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	session, err := s.newAuthSession(ctx, user, dto.IpAddr, dto.DeviceInfo, false)
	if err != nil {
		return nil, err
	}

	if err = s.otpRepo.Delete(ctx, otp); err != nil {
		return nil, err
	}

	displayName := dto.Username
	if dto.FirstName != "" {
		displayName = dto.FirstName
		if dto.LastName != "" {
			displayName = displayName + " " + dto.LastName
		}
	}

	go func() {
		if err = s.notificationsClient.SendGreetingEmail(ctx, user.Email, displayName); err != nil {
			s.log.Error("AuthService.SignUp - notificationsClient.SendGreetingEmail", "err", err, "to", user.Email)
		}
	}()

	return &dtos.SignUpResponse{Session: session, User: user}, nil
}

func (s *AuthService) RequestEmailConfirmationCode(ctx context.Context, email string) error {
	isExists, err := s.usersRepo.CheckExistsWithEmail(ctx, email)
	if err != nil {
		return err
	}
	if isExists {
		return ErrUserAlreadyExists
	}

	otpCode, err := s.createOtp(ctx, email, 0)
	if err != nil {
		return err
	}
	if err = s.notificationsClient.SendSignUpConfirmationEmail(ctx, email, otpCode); err != nil {
		s.log.Error("Failed to send signup confirmation email with otp code", "to", email)
		return err
	}
	return nil
}

func (s *AuthService) SignIn(ctx context.Context, dto *dtos.SignInRequest) (*dtos.SignInResponse, error) {
	user, err := s.usersRepo.GetByLogin(ctx, dto.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrDeactivatedAccount
	}

	err = s.securityProvider.ComparePasswords(user.Password, dto.Password)
	if err != nil {
		s.log.Error(fmt.Errorf("AuthService.SignIn - securityProvider.ComparePasswords: %w", err))
		return nil, ErrInvalidCredentials
	}

	if user.Is2FAEnabled() {
		otpCode, err := s.createOtp(ctx, "", user.Id)
		if err != nil {
			return nil, err
		}
		contact := user.TwoFactorAuth.Contact
		switch user.TwoFactorAuth.Transport {
		case entity.TWO_FA_EMAIL:
			toEmail := user.Email
			if contact != "" {
				toEmail = contact
			}
			if err = s.notificationsClient.Send2FAEmail(ctx, toEmail, otpCode); err != nil {
				return nil, err
			}
		case entity.TWO_FA_TELEGRAM:
			if err = s.tgApi.SendTextMsg(ctx, contact, fmt.Sprintf("Authorization code: %s", otpCode)); err != nil {
				return nil, err
			}
		case entity.TWO_FA_TOTP_APP:
			return &dtos.SignInResponse{
				User:             user,
				ConfirmationCode: otpCode,
			}, nil
		default:
			return nil, ErrUnsupported2FAMethod
		}
		return &dtos.SignInResponse{User: user}, nil
	}

	session, err := s.newAuthSession(ctx, user, dto.IpAddr, dto.DeviceInfo, dto.RememberMe)
	if err != nil {
		return nil, err
	}

	return &dtos.SignInResponse{
		User:    user,
		Session: session,
	}, nil
}

func (s *AuthService) VerifyTwoFa(ctx context.Context, dto *dtos.Verify2FARequest) (*entity.AuthSession, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrOtpIsNotValid
		}
		return nil, err
	}
	otpToCompare := dto.Code
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		otpToCompare = dto.SignInConfirmationCode
	}

	err = s.securityProvider.ComparePasswords(otp.Code, otpToCompare)
	if err != nil {
		s.log.Error(fmt.Errorf("AuthService.Verify2FA - invalid otp: %w", err))
		return nil, ErrOtpIsNotValid
	}

	user, err := s.usersRepo.GetByLogin(ctx, otp.UserEmail)
	if err != nil {
		s.log.Error(fmt.Errorf("Failed to get user by email in existing OTP token: %w", err), "email", otp.UserEmail)
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrDeactivatedAccount
	}
	if !user.Is2FAEnabled() {
		return nil, Err2FANotEnabled
	}
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		decryptedSecret, err := s.securityProvider.DecryptSymmetric(user.TwoFactorAuth.TotpSecret, user.PrivateKey)
		if err != nil {
			return nil, err
		}
		isValid := s.securityProvider.ValidateTOTP(dto.Code, decryptedSecret)
		if !isValid {
			return nil, ErrOtpIsNotValid
		}
	}

	session, err := s.newAuthSession(ctx, user, dto.IpAddr, dto.DeviceInfo, dto.RememberMe)
	if err != nil {
		return nil, err
	}

	if err = s.otpRepo.Delete(ctx, otp); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) ConfirmTwoFaAddition(ctx context.Context, dto *dtos.Confirm2FARequest) (*entity.TwoFactorAuth, error) {
	twoFactorAuth := &entity.TwoFactorAuth{
		UserId:    dto.UserId,
		Transport: dto.Typ,
		Enabled:   true,
	}

	user, err := s.usersRepo.GetByID(ctx, dto.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if dto.Typ == entity.TWO_FA_TOTP_APP {
		isValid := s.securityProvider.ValidateTOTP(dto.ConfirmationCode, dto.TotpSecret)
		if !isValid {
			return nil, ErrOtpIsNotValid
		}

		encryptedSecret, err := s.securityProvider.EncryptSymmetric(dto.TotpSecret, user.PrivateKey)
		if err != nil {
			return nil, err
		}
		twoFactorAuth.TotpSecret = encryptedSecret
	} else {
		otp, err := s.otpRepo.GetByUserId(ctx, dto.UserId)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, ErrOtpIsNotValid
			}
			return nil, err
		}

		err = s.securityProvider.ComparePasswords(otp.Code, dto.ConfirmationCode)
		if err != nil {
			s.log.Error(fmt.Errorf("AuthService.ConfirmTwoFaAddition - securityProvider.ComparePasswords: %w", err))
			return nil, ErrOtpIsNotValid
		}

		if err := s.otpRepo.Delete(ctx, otp); err != nil {
			return nil, err
		}
	}

	if dto.Typ == entity.TWO_FA_EMAIL || dto.Typ == entity.TWO_FA_SMS {
		twoFactorAuth.Contact = dto.Contact
	}

	twoFactorAuth, err = s.usersRepo.CreateTwoFa(ctx, twoFactorAuth)
	if err != nil {
		return nil, err
	}

	return twoFactorAuth, nil
}

type TwoFAConnectInfo struct {
	Url        string
	TotpSecret string
}

func (s *AuthService) handleAddTwoFaEmail(ctx context.Context, user *entity.User, contact string) error {
	emailRecipient := user.Email
	if contact != "" {
		emailRecipient = contact
	}

	otpCode, err := s.createOtp(ctx, "", user.Id)
	if err != nil {
		return err
	}

	return s.notificationsClient.Send2FAEmail(ctx, emailRecipient, otpCode)
}

func (s *AuthService) handleAddTwoFaTelegram(ctx context.Context, userId int) (string, error) {
	otpCode, err := s.createOtp(ctx, "", userId)
	if err != nil {
		return "", err
	}

	msgCode := s.securityProvider.GenerateOTPCode()
	link := s.tgApi.GetStartLinkWithCode(msgCode)

	go func() {
		startTime := time.Now()
		retryTimeout := 5 * time.Second
		attemptsLeft := s.otpTTL / retryTimeout
		for attemptsLeft > 0 {
			time.Sleep(retryTimeout)
			attemptsLeft--

			msg, err := s.tgApi.GetLatestMsg(ctx)
			if err != nil {
				s.log.Error(fmt.Errorf("AuthService - handleAddTwoFaTelegram - s.tgApi.GetLatestMsg: %w", err))
			}

			msgParts := strings.Split(msg.Text, " ")
			if msg.DateSent.Before(startTime) || len(msgParts) < 2 || msgParts[1] != msgCode {
				continue
			}

			err = s.usersRepo.UpdateTwoFaContact(ctx, userId, msg.ChatId)
			if err != nil {
				s.log.Error(fmt.Errorf("AuthService - handleAddTwoFaTelegram - s.usersRepo.UpdateTwoFaContact: %w", err))
				continue
			}

			err = s.tgApi.SendTextMsg(ctx, msg.ChatId, fmt.Sprintf("Authorization code: %s", otpCode))
			if err != nil {
				s.log.Error(fmt.Errorf("AuthService - handleAddTwoFaTelegram - s.tgApi.SendTextMsg: %w", err))
				continue
			}

			return
		}
	}()

	return link, nil
}

func (s *AuthService) RequestTwoFaAddition(ctx context.Context, dto *dtos.Add2FARequest) (*TwoFAConnectInfo, error) {
	user, err := s.usersRepo.GetByID(ctx, dto.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user.Is2FAEnabled() {
		return nil, Err2FaAlreadyAdded
	}
	switch dto.Typ {
	case entity.TWO_FA_EMAIL:
		return nil, s.handleAddTwoFaEmail(ctx, user, dto.Contact)
	case entity.TWO_FA_TELEGRAM:
		link, err := s.handleAddTwoFaTelegram(ctx, user.Id)
		if err != nil {
			return nil, err
		}
		return &TwoFAConnectInfo{Url: link}, nil
	case entity.TWO_FA_TOTP_APP:
		secret := s.securityProvider.GenerateSecretTokenUrlSafe(config.TOTP_SECRET_LENGTH)
		url := s.securityProvider.GenerateTOTPEnrollUrl(user.Email, secret)
		return &TwoFAConnectInfo{Url: url, TotpSecret: secret}, nil
	default:
		return nil, ErrUnsupported2FAMethod
	}
}

func (s *AuthService) DeactivateAccount(ctx context.Context, userId int) error {
	user, err := s.usersRepo.UpdateIsActiveById(ctx, userId, false)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := s.notificationsClient.SendAccDeactivationEmail(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) GetActiveSessions(
	ctx context.Context,
	userId int,
) ([]entity.AuthSession, error) {
	sessions, err := s.sessionsRepo.GetAllByUserId(ctx, userId)
	if err != nil {
		s.log.Error(fmt.Errorf("AuthService.GetActiveSessions - sessionsRepo.GetAllByUserId: %w", err), "userId", userId)
		return nil, err
	}

	return sessions, nil
}

func (s *AuthService) DeleteSession(ctx context.Context, userId int, sessionId string) error {
	if err := s.sessionsRepo.DeleteById(ctx, userId, sessionId); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrSessionNotFound
		}
	}
	return nil
}

func (s *AuthService) PingSession(
	ctx context.Context,
	userId int,
	sessionId string,
) (*entity.AuthSession, error) {
	session, err := s.sessionsRepo.GetById(ctx, userId, sessionId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	sessionTTL := s.defaultSessionTTL
	if session.IsLongLived {
		sessionTTL = s.longLivedSessionTTL
	}

	now := time.Now()
	err = s.sessionsRepo.UpdateById(ctx, userId, sessionId, now, sessionTTL)
	if err != nil {
		return nil, err
	}
	session.LastSeenAt = now

	return session, nil
}

func (s *AuthService) ExportLoginToken(ctx context.Context, dto *dtos.ExportLoginTokenRequest) (*entity.QRCodeLoginToken, error) {
	err := s.loginTokenRepo.DeleteAllByClient(ctx, dto.ClientId)
	if err != nil {
		return nil, err
	}

	tokenValue := s.securityProvider.GenerateSecretTokenUrlSafe(config.LOGIN_TOKEN_LENGTH)
	token := &entity.QRCodeLoginToken{
		ClientId:   dto.ClientId,
		Value:      tokenValue,
		IpAddr:     dto.IpAddr,
		DeviceInfo: dto.DeviceInfo,
	}

	return token, s.loginTokenRepo.CreateWithTTL(ctx, token, s.loginTokenTTL)
}

// AcceptQRLoginToken allows to authenticate another device from an authorized one (qrcode auth)
func (s *AuthService) AcceptQRLoginToken(ctx context.Context, userId int, unauthorizedClientToken string, unauthorizedClientId string) (*entity.AuthSession, error) {
	token, err := s.loginTokenRepo.FindOne(ctx, unauthorizedClientToken, unauthorizedClientId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidLoginToken
		}
		return nil, err
	}

	user, err := s.usersRepo.GetByID(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	session, err := s.newAuthSession(ctx, user, token.IpAddr, token.DeviceInfo, true)
	if err != nil {
		return nil, err
	}

	if err := s.loginTokenRepo.DeleteAllByClient(ctx, token.ClientId); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) BeginPasskeyRegistration(ctx context.Context, userId int) (gateways.WebAuthnRegistrationOptions, error) {
	user, err := s.usersRepo.GetByIDWithPasskeyCredentials(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	registrationOptions, passkeySession, err := s.webAuthnProvider.GenerateRegistrationOptions(user)
	if err != nil {
		return nil, err
	}
	if err := s.passkeySessionsRepo.Create(ctx, passkeySession); err != nil {
		return nil, err
	}

	return registrationOptions, nil
}

func (s *AuthService) FinishPasskeyRegistration(ctx context.Context, userId int, rawCredential []byte) error {
	passkeySession, err := s.passkeySessionsRepo.GetByUserId(ctx, userId)
	if err != nil {
		return err
	}

	cred, err := s.webAuthnProvider.VerifyRegistrationOptions(userId, rawCredential, passkeySession)
	if err != nil {
		if errors.Is(err, gateways.ErrInvalidCredential) {
			return ErrInvalidPasskeyCredential
		}
		return err
	}

	if err := s.usersRepo.CreatePasskeyCredential(ctx, userId, cred); err != nil {
		return err
	}

	return nil
}
