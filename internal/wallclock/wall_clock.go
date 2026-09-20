// Package wallclock answers one question: "what time is it right now?"
//
// Tests can pretend time is frozen. The live website uses the real clock.
package wallclock

import "time"

// WallClock is a clock that other packages can trust.
type WallClock interface {
	WhatTimeIsItRightNow() time.Time
}

// RealWallClock looks at the computer's real clock.
type RealWallClock struct{}

// WhatTimeIsItRightNow returns the computer's current time.
func (realWallClock RealWallClock) WhatTimeIsItRightNow() time.Time {
	return time.Now().UTC()
}

// FrozenWallClock always answers with the same time.
// Tests use this so dates never surprise them.
type FrozenWallClock struct {
	FrozenMoment time.Time
}

// WhatTimeIsItRightNow returns the frozen moment.
func (frozenWallClock FrozenWallClock) WhatTimeIsItRightNow() time.Time {
	return frozenWallClock.FrozenMoment
}
