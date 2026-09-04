package authtoken

import (
	"testing"

	"uuid"
)

func TestSignedTokenRoundTrip(t *testing.T) {
	userID := uuid.New()

	token, err := NewSignedToken(userID)
	if err != nil {
		t.Fatalf("NewSignedToken() error = %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("claims.UserID = %v, want %v", claims.UserID, userID)
	}
}
