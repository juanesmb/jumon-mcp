package security

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// PeekUnverifiedClaims is debug-only and should never log full tokens.
func PeekUnverifiedClaims(tokenString string) string {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return fmt.Sprintf("parse unverified failed: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "claims are not map claims"
	}
	iss, _ := claims["iss"].(string)
	sub, _ := claims["sub"].(string)
	return fmt.Sprintf("iss=%q aud=%v sub=%q scp=%v scope=%v", iss, claims["aud"], sub, claims["scp"], claims["scope"])
}
