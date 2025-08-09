package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"GopherTales/internal/database"
	"GopherTales/internal/models"
)

type UserService struct {
	db *database.MongoDB
}

func NewUserService(db *database.MongoDB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Register(name, email, password string) (*models.User, error) {
	ctx := context.Background()

	// Check if user exists
	var existing models.User
	err := s.db.Database.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&existing)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Progress:     make(map[string]int),
		Bookmarks:    []models.Bookmark{},
	}

	result, err := s.db.Database.Collection("users").InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (s *UserService) Login(email, password string) (*models.User, error) {
	ctx := context.Background()

	var user models.User
	err := s.db.Database.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func (s *UserService) UpdateProgress(userID primitive.ObjectID, gopher string, progress int) error {
	ctx := context.Background()

	update := bson.M{
		"$set": bson.M{
			"progress." + gopher: progress,
			"updated_at":         time.Now(),
		},
	}

	_, err := s.db.Database.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

func (s *UserService) AddBookmark(userID primitive.ObjectID, bookmark models.Bookmark) error {
	ctx := context.Background()

	update := bson.M{
		"$push": bson.M{"bookmarks": bookmark},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	_, err := s.db.Database.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update)
	return err
}

func (s *UserService) GetUserByID(userID primitive.ObjectID) (*models.User, error) {
	ctx := context.Background()

	var user models.User
	err := s.db.Database.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
