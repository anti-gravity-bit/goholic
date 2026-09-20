// Package writerpassport is the door key for the writer dashboard.
//
// Readers do not need a key.
// Only the writer who knows the username and password gets a signed cookie.
package writerpassport

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/anti-gravity-bit/goholic/internal/wallclock"
)

const writerSessionCookieName = "goholic_writer_session"
const writerSessionLifetime = 12 * time.Hour

// WriterPassport checks the writer's password and signs a cookie.
type WriterPassport struct {
	writerUsername              string
	hashedWriterPassword        []byte
	secretBytesUsedToSignCookie []byte
	wallClock                   wallclock.WallClock
}

// NewWriterPassport builds a door key from environment values.
func NewWriterPassport(
	writerUsername string,
	plainWriterPassword string,
	secretUsedToSignCookie string,
	wallClock wallclock.WallClock,
) (*WriterPassport, error) {
	cleanedUsername := strings.TrimSpace(writerUsername)
	if cleanedUsername == "" {
		return nil, errors.New("the writer username cannot be empty")
	}

	if strings.TrimSpace(plainWriterPassword) == "" {
		return nil, errors.New("the writer password cannot be empty")
	}

	if len(secretUsedToSignCookie) < 16 {
		return nil, errors.New("the cookie signing secret must be at least 16 characters")
	}

	hashedWriterPassword, hashError := bcrypt.GenerateFromPassword(
		[]byte(plainWriterPassword),
		bcrypt.DefaultCost,
	)
	if hashError != nil {
		return nil, hashError
	}

	return &WriterPassport{
		writerUsername:              cleanedUsername,
		hashedWriterPassword:        hashedWriterPassword,
		secretBytesUsedToSignCookie: []byte(secretUsedToSignCookie),
		wallClock:                   wallClock,
	}, nil
}

// WriterSessionCookieName is the name of the signed cookie.
func WriterSessionCookieName() string {
	return writerSessionCookieName
}

// TryToOpenTheWriterDoor checks the username and password.
func (writerPassport *WriterPassport) TryToOpenTheWriterDoor(
	typedUsername string,
	typedPassword string,
) (string, error) {
	if typedUsername != writerPassport.writerUsername {
		return "", errors.New("that username or password is wrong")
	}

	passwordError := bcrypt.CompareHashAndPassword(writerPassport.hashedWriterPassword, []byte(typedPassword))
	if passwordError != nil {
		return "", errors.New("that username or password is wrong")
	}

	return writerPassport.buildSignedSessionToken(), nil
}

// DoesThisSignedSessionStillCount checks a cookie from the browser.
func (writerPassport *WriterPassport) DoesThisSignedSessionStillCount(
	signedSessionToken string,
) bool {
	tokenPieces := strings.Split(signedSessionToken, ".")
	if len(tokenPieces) != 2 {
		return false
	}

	unixSecondsAsText := tokenPieces[0]
	offeredSignature := tokenPieces[1]
	expectedSignature := writerPassport.signUnixSeconds(unixSecondsAsText)
	if !hmac.Equal([]byte(offeredSignature), []byte(expectedSignature)) {
		return false
	}

	unixSeconds, parseError := strconv.ParseInt(unixSecondsAsText, 10, 64)
	if parseError != nil {
		return false
	}

	sessionCreatedAt := time.Unix(unixSeconds, 0).UTC()
	rightNow := writerPassport.wallClock.WhatTimeIsItRightNow()
	if rightNow.Sub(sessionCreatedAt) > writerSessionLifetime {
		return false
	}

	if sessionCreatedAt.After(rightNow.Add(time.Minute)) {
		return false
	}

	return true
}

func (writerPassport *WriterPassport) buildSignedSessionToken() string {
	unixSecondsAsText := fmt.Sprintf("%d", writerPassport.wallClock.WhatTimeIsItRightNow().Unix())
	return unixSecondsAsText + "." + writerPassport.signUnixSeconds(unixSecondsAsText)
}

func (writerPassport *WriterPassport) signUnixSeconds(unixSecondsAsText string) string {
	messageAuthentication := hmac.New(sha256.New, writerPassport.secretBytesUsedToSignCookie)
	_, _ = messageAuthentication.Write([]byte(unixSecondsAsText))
	return hex.EncodeToString(messageAuthentication.Sum(nil))
}
