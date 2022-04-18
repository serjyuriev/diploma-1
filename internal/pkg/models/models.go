package models

import (
	"time"

	"github.com/golang-jwt/jwt"
)

type Balance struct {
	Current   Point
	Withdrawn Point
}

type Claims struct {
	jwt.StandardClaims
	Username string `json:"username"`
}

type Order struct {
	Number      string
	Status      string
	Sum         Point
	UploadedAt  time.Time
	ProcessedAt time.Time
}

type Point int64

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
