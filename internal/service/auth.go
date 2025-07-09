package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"shortener/internal/model"
	"shortener/internal/shared/logger"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

const cookieTokenName string = "token"

type userIDKey struct{}

var (
	ErrNotValid error = errors.New("token is not valid")
	ErrEmpty    error = errors.New("token is empty")
)

type authService struct {
	log         *logger.Logger
	secret      []byte
	expireAfter time.Duration
}

func NewAuthService(log *logger.Logger, secret []byte, expireAfter time.Duration) *authService {
	return &authService{
		log:         log,
		secret:      secret,
		expireAfter: expireAfter,
	}
}

func (s *authService) UserIDFromContext(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(userIDKey{}).(string)
	return u, ok
}

func (s *authService) CheckInMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := s.getOrCreateUserID(w, r)
		if err != nil {
			s.writeInternalError(w, err, "get or create user", "auth.CheckInMiddleware")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *authService) getOrCreateUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieTokenName)
	if err != nil {
		return s.newUser(w)
	}
	return s.userID(cookie.Value)
}

func (s *authService) newUser(w http.ResponseWriter) (string, error) {
	userUUID, err := uuid.NewV4()
	if err != nil {
		s.writeInternalError(w, err, "new uuid for user_id", "auth.newUser")
		return "", err
	}

	userID := userUUID.String()
	tokenString, err := s.createToken(userID)
	if err != nil {
		s.writeInternalError(w, err, "create token", "auth.newUser")
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieTokenName,
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		// Secure:   true,
		// SameSite: http.SameSiteStrictMode,
		SameSite: http.SameSiteNoneMode,
	})

	return userID, nil
}

func (s *authService) createToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, model.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expireAfter)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("unable to create string token: %w", err)
	}

	return tokenString, nil
}

func (s *authService) userID(tokenString string) (string, error) {
	c := model.Claims{}
	token, err := s.parseToken(tokenString, &c)
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrNotValid
	}

	return c.UserID, nil
}

func (s *authService) parseToken(tokenString string, c *model.Claims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, c,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return s.secret, nil
		})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *authService) writeInternalError(
	w http.ResponseWriter,
	err error,
	reason string,
	op string,
) {
	s.log.Error(
		reason,
		logger.String("op", op),
		logger.Error(err),
	)
	http.Error(w, "internal error", http.StatusInternalServerError)
}
