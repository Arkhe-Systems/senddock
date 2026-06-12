package service

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/pquerna/otp/totp"
)

const (
	totpIssuer       = "SendDock"
	recoveryCodeBits = 40
	recoveryCodeN    = 10
)

type TOTPSetup struct {
	Secret        string
	OtpauthURL    string
	RecoveryCodes []string
}

func GenerateTOTPSetup(accountEmail string) (TOTPSetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: accountEmail,
	})
	if err != nil {
		return TOTPSetup{}, err
	}

	codes, err := generateRecoveryCodes(recoveryCodeN)
	if err != nil {
		return TOTPSetup{}, err
	}

	return TOTPSetup{
		Secret:        key.Secret(),
		OtpauthURL:    key.URL(),
		RecoveryCodes: codes,
	}, nil
}

func ValidateTOTPCode(secret, code string) bool {
	return totp.Validate(strings.TrimSpace(code), secret)
}

func generateRecoveryCodes(n int) ([]string, error) {
	out := make([]string, n)
	for i := range n {
		buf := make([]byte, recoveryCodeBits/8)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		raw := strings.ToUpper(hex.EncodeToString(buf))
		out[i] = raw[:5] + "-" + raw[5:]
	}
	return out, nil
}

func normalizeRecoveryCode(code string) string {
	out := strings.ToUpper(strings.TrimSpace(code))
	out = strings.ReplaceAll(out, "-", "")
	out = strings.ReplaceAll(out, " ", "")
	return out
}

func IsLikelyRecoveryCode(input string) bool {
	clean := normalizeRecoveryCode(input)
	if len(clean) != 10 {
		return false
	}
	for _, c := range clean {
		isHex := (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')
		if !isHex {
			return false
		}
	}
	return true
}

var _ = base32.StdEncoding
var ErrInvalidTOTPCode = errors.New("invalid 2FA code")
var ErrInvalidRecoveryCode = errors.New("invalid recovery code")
