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
	"github.com/modulix-systems/goose-talk/logger"
)

func (s *Service) SignUp(
	ctx context.Context,
	dto *dtos.SignUpRequest,
) (*dtos.SignUpResponse, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.SignUp"
	log := s.log.With("op", op, "correlationId", correlationId, "email", dto.Email)
	start := time.Now()
	defer func() { log.Debug("SignUp finished", "duration", time.Since(start)) }()

	log.Debug("Signing up", "username", dto.Username, "firstName", dto.FirstName, "lastName", dto.LastName, "birthDate", dto.BirthDate, "ip", dto.IpAddr, "confirmationCode", dto.ConfirmationCode, "deviceInfo", dto.DeviceInfo, "photoUrl", dto.PhotoUrl)
	userExists, err := s.usersRepo.CheckExistsWithEmail(ctx, dto.Email)
	if err != nil {
		log.Error("error checking user existence", "err", err)
		return nil, fmt.Errorf("%s - error checking user existence: %w", op, err)
	}
	log.Debug("checked user existence", "exists", userExists)
	if userExists {
		return nil, ErrUserAlreadyExists
	}

	if dto.ConfirmationCode == "" {
		otpCode, err := s.createOtp(ctx, dto.Email, 0)
		if err != nil {
			return nil, fmt.Errorf("%s - error creating otp: %w", op, err)
		}
		if err = s.notificationsClient.SendEmailVerifyEmail(ctx, dto.Email, dto.Username, otpCode); err != nil {
			log.Error("failed to send verification email", "err", err, "to", dto.Email)
			return nil, fmt.Errorf("%s - error sending email verification email: %w", op, err)
		}
		return nil, ErrEmailUnverified
	}

	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrOtpIsNotValid
		}
		return nil, err
	}

	log.Debug("comparing otp codes", "otpPresent", otp != nil)
	err = s.securityProvider.ComparePasswords(otp.Code, dto.ConfirmationCode)
	if err != nil {
		log.Error("invalid otp", "err", err, "otpUserEmail", otp.UserEmail)
		return nil, ErrOtpIsNotValid
	}

	log.Debug("hashing password", "pwdLen", len(dto.Password))
	hashedPassword, err := s.securityProvider.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.usersRepo.Save(
		ctx,
		&entity.User{
			FirstName:  dto.FirstName,
			LastName:   dto.LastName,
			Username:   dto.Username,
			BirthDate:  dto.BirthDate,
			Email:      dto.Email,
			AboutMe:    dto.AboutMe,
			Password:   hashedPassword,
			PrivateKey: s.securityProvider.GeneratePrivateKey(),
		},
	)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		log.Error("failed to save user", "err", err, "email", dto.Email)
		return nil, err
	}
	log.Debug("user saved", "userId", user.Id, "email", user.Email)

	session, err := s.newAuthSession(ctx, user, dto.IpAddr, dto.DeviceInfo, false, true)
	if err != nil {
		return nil, err
	}
	log.Debug("created auth session", "userId", user.Id, "sessionId", session.Id)

	if err = s.otpRepo.Delete(ctx, otp); err != nil {
		log.Error("failed to delete otp after signup", "err", err, "otpUserEmail", otp.UserEmail)
		return nil, err
	}

	if err = s.notificationsClient.SendSignUpEmail(ctx, user); err != nil {
		log.Error("failed to send signup email", "err", err, "to", user.Email)
	} else {
		log.Debug("signup email sent", "to", user.Email)
	}

	return &dtos.SignUpResponse{Session: session, User: user}, nil
}

func (s *Service) SignIn(ctx context.Context, dto *dtos.SignInRequest) (*dtos.SignInResponse, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.SignIn"
	log := s.log.With("op", op, "correlationId", correlationId, "login", dto.Login)
	start := time.Now()
	defer func() { log.Debug("SignIn finished", "duration", time.Since(start)) }()

	user, err := s.usersRepo.GetByLogin(ctx, dto.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	log.Debug("fetched user by login", "userId", user.Id, "isActive", user.IsActive)
	if !user.IsActive {
		return nil, ErrDeactivatedAccount
	}

	log.Debug("comparing password hash", "userId", user.Id)
	err = s.securityProvider.ComparePasswords(user.Password, dto.Password)
	if err != nil {
		log.Error("invalid password", "err", err, "login", dto.Login, "userId", user.Id)
		return nil, ErrInvalidCredentials
	}

	if user.Is2FAEnabled() {
		log.Debug("user has 2FA enabled", "userId", user.Id, "method", user.TwoFactorAuth.Method)
		otpCode, err := s.createOtp(ctx, "", user.Id)
		if err != nil {
			return nil, err
		}
		contact := user.TwoFactorAuth.Contact
		switch user.TwoFactorAuth.Method {
		case entity.TWO_FA_EMAIL:
			toEmail := user.Email
			if contact != "" {
				toEmail = contact
			}
			log.Debug("sending 2fa email", "to", toEmail, "userId", user.Id)
			if err = s.notificationsClient.SendConfirmEmailTwoFaEmail(ctx, toEmail, user.GetDisplayName(), otpCode, user.Language); err != nil {
				log.Error("failed to send 2fa email", "err", err, "to", toEmail, "userId", user.Id)
				return nil, err
			}
		case entity.TWO_FA_TELEGRAM:
			log.Debug("sending 2fa telegram message", "contact", contact, "userId", user.Id)
			if err = s.tgApi.SendTextMsg(ctx, contact, fmt.Sprintf("Authorization code: %s", otpCode)); err != nil {
				log.Error("failed to send 2fa telegram message", "err", err, "contact", contact, "userId", user.Id)
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

	session, err := s.newAuthSession(ctx, user, dto.IpAddr, dto.DeviceInfo, dto.RememberMe, false)
	if err != nil {
		return nil, err
	}
	log.Debug("created auth session", "userId", user.Id, "sessionId", session.Id)

	return &dtos.SignInResponse{
		User:    user,
		Session: session,
	}, nil
}

func (s *Service) VerifyTwoFa(ctx context.Context, dto *dtos.Verify2FARequest) (*entity.AuthSession, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.VerifyTwoFa"
	log := s.log.With("op", op, "correlationId", correlationId, "email", dto.Email)
	start := time.Now()
	defer func() { log.Debug("VerifyTwoFa finished", "duration", time.Since(start)) }()

	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrOtpIsNotValid
		}
		return nil, err
	}
	log.Debug("fetched otp for email", "email", dto.Email, "otpUserEmail", otp.UserEmail)
	otpToCompare := dto.Code
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		otpToCompare = dto.SignInConfirmationCode
	}

	log.Debug("comparing otp for verify twofa", "email", dto.Email)
	err = s.securityProvider.ComparePasswords(otp.Code, otpToCompare)
	if err != nil {
		log.Error("invalid otp", "err", err, "email", dto.Email)
		return nil, ErrOtpIsNotValid
	}

	user, err := s.usersRepo.GetByLogin(ctx, otp.UserEmail)
	if err != nil {
		log.Error("failed to get user by email", "err", err, "email", otp.UserEmail)
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	log.Debug("verify twofa fetched user", "userId", user.Id)
	if !user.IsActive {
		return nil, ErrDeactivatedAccount
	}
	if !user.Is2FAEnabled() {
		return nil, Err2FANotEnabled
	}
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		log.Debug("validating totp app code", "userId", user.Id)
		decryptedSecret, err := s.securityProvider.DecryptSymmetric(user.TwoFactorAuth.TotpSecret, user.PrivateKey)
		if err != nil {
			log.Error("failed to decrypt totp secret", "err", err, "userId", user.Id)
			return nil, err
		}
		isValid := s.securityProvider.ValidateTOTP(dto.Code, decryptedSecret)
		if !isValid {
			log.Error("invalid totp code", "userId", user.Id)
			return nil, ErrOtpIsNotValid
		}
		log.Debug("totp code validated", "userId", user.Id)
	}

	session, err := s.newAuthSession(ctx, user, dto.IpAddr, dto.DeviceInfo, dto.RememberMe, false)
	if err != nil {
		return nil, err
	}

	if err = s.otpRepo.Delete(ctx, otp); err != nil {
		log.Error("failed to delete otp after verify twofa", "err", err, "email", otp.UserEmail)
		return nil, err
	}

	return session, nil
}

func (s *Service) CompleteAddingTwoFa(ctx context.Context, dto *dtos.Confirm2FARequest) (*entity.TwoFactorAuth, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.CompleteAddingTwoFa"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", dto.UserId)
	start := time.Now()
	defer func() { log.Debug("CompleteAddingTwoFa finished", "duration", time.Since(start)) }()

	twoFactorAuth := &entity.TwoFactorAuth{
		UserId:  dto.UserId,
		Method:  dto.Typ,
		Enabled: true,
	}

	user, err := s.usersRepo.GetByID(ctx, dto.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	log.Debug("fetched user for adding 2fa", "userId", user.Id)

	if dto.Typ == entity.TWO_FA_TOTP_APP {
		log.Debug("validating totp during complete add 2fa", "userId", user.Id)
		isValid := s.securityProvider.ValidateTOTP(dto.ConfirmationCode, dto.TotpSecret)
		if !isValid {
			log.Error("invalid totp during complete add 2fa", "userId", user.Id)
			return nil, ErrOtpIsNotValid
		}

		encryptedSecret, err := s.securityProvider.EncryptSymmetric(dto.TotpSecret, user.PrivateKey)
		if err != nil {
			log.Error("failed to encrypt totp secret", "err", err, "userId", user.Id)
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
		log.Debug("fetched otp for user during complete add 2fa", "userId", dto.UserId)

		err = s.securityProvider.ComparePasswords(otp.Code, dto.ConfirmationCode)
		if err != nil {
			log.Error("invalid confirmation code during complete add 2fa", "err", err, "userId", dto.UserId)
			return nil, ErrOtpIsNotValid
		}

		if err := s.otpRepo.Delete(ctx, otp); err != nil {
			log.Error("failed to delete otp after completing add 2fa", "err", err, "userId", dto.UserId)
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

func (s *Service) handleAddTwoFaEmail(ctx context.Context, user *entity.User, contact string) error {
	to := user.Email
	if contact != "" {
		to = contact
	}

	otpCode, err := s.createOtp(ctx, "", user.Id)
	if err != nil {
		return err
	}

	return s.notificationsClient.SendConfirmEmailTwoFaEmail(ctx, to, user.GetDisplayName(), otpCode, user.Language)
}

func (s *Service) handleAddTwoFaTelegram(ctx context.Context, userId int) (string, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.handleAddTwoFaTelegram"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId)
	start := time.Now()
	defer func() { log.Debug("handleAddTwoFaTelegram finished", "duration", time.Since(start)) }()

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
		log.Debug("starting telegram 2fa background listener", "attempts", attemptsLeft)
		for attemptsLeft > 0 {
			time.Sleep(retryTimeout)
			attemptsLeft--

			msg, err := s.tgApi.GetLatestMsg(ctx)
			if err != nil {
				log.Error("failed to get latest telegram message", "err", err, "attemptsLeft", attemptsLeft)
				continue
			}

			msgParts := strings.Split(msg.Text, " ")
			if msg.DateSent.Before(startTime) || len(msgParts) < 2 || msgParts[1] != msgCode {
				log.Debug("telegram msg did not match code or is old", "msgDate", msg.DateSent, "attemptsLeft", attemptsLeft)
				continue
			}

			log.Debug("telegram code matched, updating twofa contact", "chatId", msg.ChatId, "userId", userId)
			err = s.usersRepo.UpdateTwoFaContact(ctx, userId, msg.ChatId)
			if err != nil {
				log.Error("failed to update 2fa contact", "err", err, "chatId", msg.ChatId)
				continue
			}

			err = s.tgApi.SendTextMsg(ctx, msg.ChatId, fmt.Sprintf("Authorization code: %s", otpCode))
			if err != nil {
				log.Error("failed to send telegram otp", "err", err, "chatId", msg.ChatId)
				continue
			}

			log.Debug("telegram 2fa flow completed for user", "userId", userId, "chatId", msg.ChatId)
			return
		}
		log.Debug("telegram 2fa background listener finished without match", "userId", userId)
	}()

	return link, nil
}

func (s *Service) RequestAddingTwoFa(ctx context.Context, dto *dtos.Add2FARequest) (*TwoFAConnectInfo, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.RequestAddingTwoFa"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", dto.UserId)
	start := time.Now()
	defer func() { log.Debug("RequestAddingTwoFa finished", "duration", time.Since(start)) }()

	user, err := s.usersRepo.GetByID(ctx, dto.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		log.Error("failed to get user", "err", err, "userId", dto.UserId)
		return nil, err
	}
	log.Debug("fetched user for request adding 2fa", "userId", user.Id, "is2faEnabled", user.Is2FAEnabled())
	if user.Is2FAEnabled() {
		return nil, Err2FaAlreadyAdded
	}
	switch dto.Typ {
	case entity.TWO_FA_EMAIL:
		if err := s.handleAddTwoFaEmail(ctx, user, dto.Contact); err != nil {
			log.Error("failed to request 2fa email", "err", err, "userId", dto.UserId)
			return nil, err
		}
		return nil, nil
	case entity.TWO_FA_TELEGRAM:
		link, err := s.handleAddTwoFaTelegram(ctx, user.Id)
		if err != nil {
			log.Error("failed to request 2fa telegram", "err", err, "userId", dto.UserId)
			return nil, err
		}
		return &TwoFAConnectInfo{Url: link}, nil
	case entity.TWO_FA_TOTP_APP:
		secret := s.securityProvider.GenerateSecretTokenUrlSafe(config.TOTP_SECRET_LENGTH)
		url := s.securityProvider.GenerateTOTPEnrollUrl(user.Email, secret)
		return &TwoFAConnectInfo{Url: url, TotpSecret: secret}, nil
	default:
		log.Error("unsupported 2fa method", "method", dto.Typ)
		return nil, ErrUnsupported2FAMethod
	}
}

func (s *Service) DeactivateAccount(ctx context.Context, userId int) error {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.DeactivateAccount"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId)
	start := time.Now()
	defer func() { log.Debug("DeactivateAccount finished", "duration", time.Since(start)) }()

	user, err := s.usersRepo.UpdateIsActiveById(ctx, userId, false)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrUserNotFound
		}
		log.Error("failed to deactivate account", "err", err, "userId", userId)
		return err
	}
	log.Debug("account deactivated", "userId", userId, "email", user.Email)
	if err := s.notificationsClient.SendAccountDeactivatedEmail(ctx, user.Email, user.GetDisplayName(), user.Language); err != nil {
		log.Error("failed to send account deactivated email", "err", err, "email", user.Email)
		return err
	}
	log.Debug("account deactivated email sent", "userId", userId, "email", user.Email)
	return nil
}

func (s *Service) GetActiveSessions(
	ctx context.Context,
	userId int,
) ([]entity.AuthSession, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.GetActiveSessions"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId)
	start := time.Now()
	defer func() { log.Debug("GetActiveSessions finished", "duration", time.Since(start)) }()

	sessions, err := s.sessionsRepo.GetAllByUserId(ctx, userId)
	if err != nil {
		log.Error("failed to get active sessions", "err", err)
		return nil, err
	}
	log.Debug("fetched active sessions", "userId", userId, "count", len(sessions))

	return sessions, nil
}

func (s *Service) DeleteSession(ctx context.Context, userId int, sessionId string) error {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.DeleteSession"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId, "sessionId", sessionId)
	start := time.Now()
	defer func() { log.Debug("DeleteSession finished", "duration", time.Since(start)) }()

	if err := s.sessionsRepo.DeleteById(ctx, userId, sessionId); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrSessionNotFound
		}
		log.Error("failed to delete session", "err", err, "userId", userId, "sessionId", sessionId)
		return err
	}
	log.Debug("deleted session", "userId", userId, "sessionId", sessionId)
	return nil
}

func (s *Service) PingSession(
	ctx context.Context,
	userId int,
	sessionId string,
) (*entity.AuthSession, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.PingSession"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId, "sessionId", sessionId)
	start := time.Now()
	defer func() { log.Debug("PingSession finished", "duration", time.Since(start)) }()

	session, err := s.sessionsRepo.GetById(ctx, userId, sessionId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		log.Error("failed to get session", "err", err, "userId", userId, "sessionId", sessionId)
		return nil, err
	}
	log.Debug("fetched session", "userId", userId, "sessionId", sessionId, "lastSeenAt", session.LastSeenAt)

	sessionTTL := s.defaultSessionTTL
	if session.IsLongLived {
		sessionTTL = s.longLivedSessionTTL
	}

	now := time.Now()
	err = s.sessionsRepo.UpdateById(ctx, userId, sessionId, now, sessionTTL)
	if err != nil {
		log.Error("failed to update session", "err", err, "userId", userId, "sessionId", sessionId)
		return nil, err
	}
	log.Debug("updated session last seen", "userId", userId, "sessionId", sessionId, "newLastSeen", now)
	session.LastSeenAt = now

	return session, nil
}

func (s *Service) ExportLoginToken(ctx context.Context, dto *dtos.ExportLoginTokenRequest) (*entity.QRCodeLoginToken, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.ExportLoginToken"
	log := s.log.With("op", op, "correlationId", correlationId, "clientId", dto.ClientId)
	start := time.Now()
	defer func() { log.Debug("ExportLoginToken finished", "duration", time.Since(start)) }()

	err := s.loginTokenRepo.DeleteAllByClient(ctx, dto.ClientId)
	if err != nil {
		log.Error("failed to delete login tokens", "err", err, "clientId", dto.ClientId)
		return nil, err
	}
	log.Debug("deleted existing login tokens for client", "clientId", dto.ClientId)

	tokenValue := s.securityProvider.GenerateSecretTokenUrlSafe(config.LOGIN_TOKEN_LENGTH)
	token := &entity.QRCodeLoginToken{
		ClientId:   dto.ClientId,
		Value:      tokenValue,
		IpAddr:     dto.IpAddr,
		DeviceInfo: dto.DeviceInfo,
	}

	if err := s.loginTokenRepo.CreateWithTTL(ctx, token, s.loginTokenTTL); err != nil {
		log.Error("failed to create login token", "err", err, "clientId", dto.ClientId)
		return nil, err
	}
	log.Debug("created login token", "clientId", dto.ClientId, "tokenIdPresent", token != nil)

	return token, nil
}

// AcceptQRLoginToken allows to authenticate another device from an authorized one (qrcode auth)
func (s *Service) AcceptQRLoginToken(ctx context.Context, userId int, unauthorizedClientToken string, unauthorizedClientId string) (*entity.AuthSession, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.AcceptQRLoginToken"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId, "unauthorizedClientId", unauthorizedClientId)
	start := time.Now()
	defer func() { log.Debug("AcceptQRLoginToken finished", "duration", time.Since(start)) }()

	token, err := s.loginTokenRepo.FindOne(ctx, unauthorizedClientToken, unauthorizedClientId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidLoginToken
		}
		log.Error("failed to find login token", "err", err, "unauthorizedClientId", unauthorizedClientId)
		return nil, err
	}
	log.Debug("found login token", "clientId", token.ClientId, "ip", token.IpAddr)

	user, err := s.usersRepo.GetByID(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		log.Error("failed to get user", "err", err, "userId", userId)
		return nil, err
	}
	log.Debug("fetched user for qr accept", "userId", userId)

	session, err := s.newAuthSession(ctx, user, token.IpAddr, token.DeviceInfo, true, false)
	if err != nil {
		log.Error("failed to create auth session", "err", err)
		return nil, err
	}
	log.Debug("created auth session", "userId", user.Id, "sessionId", session.Id)

	if err := s.loginTokenRepo.DeleteAllByClient(ctx, token.ClientId); err != nil {
		log.Error("failed to delete login tokens", "err", err, "clientId", token.ClientId)
		return nil, err
	}
	log.Debug("deleted login tokens after accept", "clientId", token.ClientId)

	return session, nil
}

func (s *Service) RequestPasskeyRegistrationOptions(ctx context.Context, userId int) (gateways.WebAuthnRegistrationOptions, error) {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.RequestPasskeyRegistrationOptions"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId)
	start := time.Now()
	defer func() { log.Debug("RequestPasskeyRegistrationOptions finished", "duration", time.Since(start)) }()

	user, err := s.usersRepo.GetByIDWithPasskeyCredentials(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		log.Error("failed to get user", "err", err, "userId", userId)
		return nil, err
	}
	log.Debug("fetched user for passkey registration options", "userId", userId, "existingCreds", len(user.PasskeyCredentials))
	registrationOptions, passkeySession, err := s.webAuthnProvider.GenerateRegistrationOptions(user)
	if err != nil {
		log.Error("failed to generate registration options", "err", err, "userId", userId)
		return nil, err
	}
	if err := s.passkeySessionsRepo.Create(ctx, passkeySession); err != nil {
		log.Error("failed to store passkey session", "err", err, "userId", userId)
		return nil, err
	}
	log.Debug("created passkey session stored", "userId", userId)

	return registrationOptions, nil
}

func (s *Service) CompletePasskeyRegistration(ctx context.Context, userId int, rawCredential []byte) error {
	correlationId := logger.CorrelationIDFromContext(ctx)
	op := "auth.Service.CompletePasskeyRegistration"
	log := s.log.With("op", op, "correlationId", correlationId, "userId", userId)
	start := time.Now()
	defer func() { log.Debug("CompletePasskeyRegistration finished", "duration", time.Since(start)) }()

	passkeySession, err := s.passkeySessionsRepo.GetByUserId(ctx, userId)
	if err != nil {
		log.Error("failed to get passkey session", "err", err, "userId", userId)
		return err
	}
	log.Debug("fetched passkey session", "userId", userId)

	cred, err := s.webAuthnProvider.VerifyRegistrationOptions(userId, rawCredential, passkeySession)
	if err != nil {
		if errors.Is(err, gateways.ErrInvalidCredential) {
			log.Error("invalid passkey credential", "err", err, "userId", userId)
			return ErrInvalidPasskeyCredential
		}
		log.Error("failed to verify registration options", "err", err, "userId", userId)
		return err
	}
	log.Debug("passkey credential verified", "userId", userId)

	if err := s.usersRepo.CreatePasskeyCredential(ctx, userId, cred); err != nil {
		log.Error("failed to create passkey credential", "err", err, "userId", userId)
		return err
	}
	log.Debug("passkey credential created", "userId", userId)

	return nil
}
