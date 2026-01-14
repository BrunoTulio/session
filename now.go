package session

import "time"

type TimeFunc func() time.Time

var now TimeFunc = time.Now

