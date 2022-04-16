package models

import "time"

type Order struct {
	Number      string
	Status      string
	Sum         Point
	UploadedAt  time.Time
	ProcessedAt time.Time
}

type Point int64

type User struct {
	Login    string
	Password string
}
