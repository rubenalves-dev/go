package usecase

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"raiiaa.dev/services/auth-service/internal/domain"
	"raiiaa.dev/services/auth-service/internal/ports"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	repo      ports.UserRepository
	signerKey *rsa.PrivateKey
	issuer    string
	expiry    time.Duration
}

func NewAuthService(repo ports.UserRepository, signerKey *rsa.PrivateKey, issuer string, expiry time.Duration) *AuthService {
	return &AuthService{
		repo:      repo,
		signerKey: signerKey,
		issuer:    issuer,
		expiry:    expiry,
	}
}

func (a *AuthService) Signup(ctx context.Context, email, password string) (string, *domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return "", nil, errors.New("email and password are required")
	}

	_, err := a.repo.FindByEmail(ctx, email)
	if err == nil {
		return "", nil, ports.ErrUserAlreadyExist
	}
	if !errors.Is(err, ports.ErrUserNotFound) {
		return "", nil, fmt.Errorf("check existing user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := a.repo.CreateUser(ctx, email, string(hash))
	if err != nil {
		return "", nil, fmt.Errorf("create user: %w", err)
	}

	token, err := a.issueToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return "", nil, errors.New("email and password are required")
	}

	user, err := a.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ports.ErrUserNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := a.issueToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (a *AuthService) issueToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"iss":   a.issuer,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(a.expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(a.signerKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
