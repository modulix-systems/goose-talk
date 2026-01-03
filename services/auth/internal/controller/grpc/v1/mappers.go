package rpc_v1

import (
	"time"

	pb "buf.build/gen/go/co3n/goose-proto/protocolbuffers/go/auth/v1"
	usersv1 "buf.build/gen/go/co3n/goose-proto/protocolbuffers/go/users/v1"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapTwoFATransport(t entity.TwoFATransport) usersv1.TwoFactorAuth_TwoFATransport {
	switch t {
	case entity.TWO_FA_TELEGRAM:
		return usersv1.TwoFactorAuth_TWO_FA_TRANSPORT_TELEGRAM
	case entity.TWO_FA_EMAIL:
		return usersv1.TwoFactorAuth_TWO_FA_TRANSPORT_EMAIL
	case entity.TWO_FA_SMS:
		return usersv1.TwoFactorAuth_TWO_FA_TRANSPORT_SMS
	case entity.TWO_FA_TOTP_APP:
		return usersv1.TwoFactorAuth_TWO_FA_TRANSPORT_TOTP
	default:
		return usersv1.TwoFactorAuth_TWO_FA_TRANSPORT_UNSPECIFIED
	}
}

func mapTwoFactorAuth(src *entity.TwoFactorAuth) *usersv1.TwoFactorAuth {
	if src == nil {
		return nil
	}

	return &usersv1.TwoFactorAuth{
		UserId:     int64(src.UserId),
		Transport:  mapTwoFATransport(src.Transport),
		Contact:    src.Contact,
		TotpSecret: src.TotpSecret,
		Enabled:    src.Enabled,
	}
}

func mapTimestamp(srcTime time.Time) *timestamppb.Timestamp {
	if srcTime.IsZero() {
		return nil
	}
	return timestamppb.New(srcTime)
}

func mapUser(src *entity.User) *usersv1.User {
	if src == nil {
		return nil
	}

	return &usersv1.User{
		Id:       int64(src.Id),
		Username: src.Username,
		Email:    src.Email,

		FirstName: src.FirstName,
		LastName:  src.LastName,
		AboutMe:   src.AboutMe,
		PhotoUrl:  src.PhotoUrl,

		CreatedAt: mapTimestamp(src.CreatedAt),
		UpdatedAt: mapTimestamp(src.UpdatedAt),
		BirthDate: mapTimestamp(src.BirthDate),

		TwoFactorAuth: mapTwoFactorAuth(src.TwoFactorAuth),
	}
}

func mapSession(src *entity.AuthSession) *pb.AuthSession {
	return &pb.AuthSession{
		Id:          src.Id,
		UserId:      int64(src.UserId),
		LastSeenAt:  mapTimestamp(src.LastSeenAt),
		CreatedAt:   mapTimestamp(src.CreatedAt),
		IsLongLived: src.IsLongLived,

		Location:   src.Location,
		IpAddr:     src.IpAddr,
		DeviceInfo: src.DeviceInfo,
	}
}
