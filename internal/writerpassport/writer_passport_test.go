package writerpassport

import (
	"testing"
	"time"

	"github.com/anti-gravity-bit/goholic/internal/wallclock"
)

func TestWriterPassportOpensTheDoorForTheRealWriterAndClosesItForAStranger(t *testing.T) {
	frozenClock := wallclock.FrozenWallClock{
		FrozenMoment: time.Date(2026, time.September, 17, 15, 0, 0, 0, time.UTC),
	}
	writerPassport, buildError := NewWriterPassport(
		"abir",
		"correct-horse-battery",
		"sixteen-or-more-chars",
		frozenClock,
	)
	if buildError != nil {
		t.Fatalf("could not build passport: %v", buildError)
	}

	signedSessionToken, openError := writerPassport.TryToOpenTheWriterDoor("abir", "correct-horse-battery")
	if openError != nil {
		t.Fatalf("the real writer should get in: %v", openError)
	}

	if !writerPassport.DoesThisSignedSessionStillCount(signedSessionToken) {
		t.Fatal("a fresh session token should still count")
	}

	_, strangerError := writerPassport.TryToOpenTheWriterDoor("abir", "wrong-password")
	if strangerError == nil {
		t.Fatal("a stranger with the wrong password must stay outside")
	}
}

func TestWriterPassportRejectsATamperedCookie(t *testing.T) {
	writerPassport, _ := NewWriterPassport(
		"abir",
		"correct-horse-battery",
		"sixteen-or-more-chars",
		wallclock.RealWallClock{},
	)

	signedSessionToken, _ := writerPassport.TryToOpenTheWriterDoor("abir", "correct-horse-battery")
	tamperedToken := signedSessionToken + "nope"

	if writerPassport.DoesThisSignedSessionStillCount(tamperedToken) {
		t.Fatal("a tampered cookie must not count")
	}
}

func TestWriterPassportRejectsAnOldCookie(t *testing.T) {
	oldClock := wallclock.FrozenWallClock{
		FrozenMoment: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	writerPassport, _ := NewWriterPassport(
		"abir",
		"correct-horse-battery",
		"sixteen-or-more-chars",
		oldClock,
	)

	signedSessionToken, _ := writerPassport.TryToOpenTheWriterDoor("abir", "correct-horse-battery")

	laterPassport := *writerPassport
	laterPassport.wallClock = wallclock.FrozenWallClock{
		FrozenMoment: time.Date(2026, time.January, 2, 1, 0, 0, 0, time.UTC),
	}

	if laterPassport.DoesThisSignedSessionStillCount(signedSessionToken) {
		t.Fatal("a cookie older than twelve hours must not count")
	}
}
