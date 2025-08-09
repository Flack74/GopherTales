package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	Progress     map[string]int     `bson:"progress" json:"progress"`
	Bookmarks    []Bookmark         `bson:"bookmarks" json:"bookmarks"`
}

type Bookmark struct {
	Gopher    string    `bson:"gopher" json:"gopher"`
	Arc       string    `bson:"arc" json:"arc"`
	Title     string    `bson:"title" json:"title"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
