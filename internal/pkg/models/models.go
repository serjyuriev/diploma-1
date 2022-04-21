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
	UserID int `json:"user_id"`
}

type Order struct {
	Number      string
	Status      string
	Accrual     Point
	Sum         Point
	UploadedAt  time.Time
	ProcessedAt time.Time
}

type Point int64

func (p Point) Float64() float64 {
	x := float64(p)
	x = x / 100
	return x
}

func ToPoints(f float64) Point {
	return Point((f * 100) + 0.5)
}

type User struct {
	ID       int    `json:"id,omitempty"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
