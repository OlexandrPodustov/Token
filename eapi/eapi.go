package eapi

import "time"

//JwtToken stores token itself and it's lifetime
type JwtToken struct {
	Token      string `json:"token"`
	TimeToLive time.Duration
}
