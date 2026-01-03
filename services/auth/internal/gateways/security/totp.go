package security

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

func (s *SecurityProvider) GenerateTOTPEnrollUrl(accountName string, secret string) string {
	secret = url.QueryEscape(secret)
	appName := url.PathEscape(s.appName)
	accountName = url.PathEscape(accountName)
	url := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", appName, accountName, secret, appName)

	return url
}

func (s *SecurityProvider) ValidateTOTP(code string, secret string) bool {
	unixTime := time.Now().UTC().Unix()
	interval := int64(30)

	counter := uint64(unixTime / interval)
	counterBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBuf, counter)

	mac := hmac.New(sha1.New, []byte(secret))
	if _, err := mac.Write(counterBuf); err != nil {
		return false
	}
	hash := mac.Sum(nil)
	// start := hash[len(hash)-1] & 0b00001111
	// end := start + 4 - 1
	// totpBytes := make([]byte, 0, 4)
	// for i := start; i < end; i++ {
	// 	totpBytes = append(totpBytes, hash[i]<<1|hash[i+1]>>7)
	// }
	// totpBytes = append(totpBytes, hash[end]<<1)

	// actualCode := binary.BigEndian.Uint32(totpBytes) % 1_000_000
	offset := hash[len(hash)-1] & 0x0f

	actualCode :=
		(uint32(hash[offset])&0x7f)<<24 |
			uint32(hash[offset+1])<<16 |
			uint32(hash[offset+2])<<8 |
			uint32(hash[offset+3])
	actualCode = actualCode % 1_000_000
	expectedCode, err := strconv.ParseUint(code, 10, 0)
	if err != nil {
		return false
	}

	return uint32(expectedCode) == actualCode
}
