// Package timefmt defines time layouts the standard library does not
// provide, so they have a single definition point.
package timefmt

// TimeHourMinuteOnly matches clock times at hour:minute precision: "15:04".
// time.TimeOnly also matches values carrying seconds.
const TimeHourMinuteOnly = "15:04"
