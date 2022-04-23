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
	ID            int64
	UserID        int
	Number        string
	Status        string
	AccrualStatus string
	Accrual       Point
	Sum           Point
	UploadedAt    time.Time
	ProcessedAt   time.Time
}

type Point int64

// Float64 convert Point type value to float64.
func (p Point) Float64() float64 {
	x := float64(p)
	x = x / 100
	return x
}

// ToPoints converts float value to Point type.
func ToPoints(f float64) Point {
	return Point((f * 100) + 0.5)
}

type User struct {
	ID       int    `json:"id,omitempty"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
