package authorizer

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/htrandev/gophermart/pkg/strutil"
	"github.com/stretchr/testify/suite"
)

type AuthorizerSuite struct {
	suite.Suite
	key        string
	authorizer *Authorizer
}

func TestAuthorizer(t *testing.T) {
	suite.Run(t, new(AuthorizerSuite))
}

func (a *AuthorizerSuite) SetupSuite() {
	a.key = "supersecretkey"
	ttl := time.Minute

	auth := New(a.key, ttl)

	a.authorizer = auth
}

func (a *AuthorizerSuite) TestHashPassword() {
	a.Run("valid", func() {
		password := "secretpassword"
		hash, err := a.authorizer.HashPassword(password)
		a.Require().NoError(err)
		a.Require().NotEqual(password, hash)
	})

	a.Run("long password", func() {
		password := strutil.Random(100)
		_, err := a.authorizer.HashPassword(password)
		a.Require().Error(err)
	})
}

func (a *AuthorizerSuite) TestValidatePassword() {
	password := "secretpassword"
	hash, err := a.authorizer.HashPassword(password)
	a.Require().NoError(err)

	a.Run("valid", func() {
		valid := a.authorizer.ValidatePassword(hash, password)
		a.Require().True(valid)
	})
	a.Run("invalid", func() {
		valid := a.authorizer.ValidatePassword(hash, "invalidpassword")
		a.Require().False(valid)
	})
}

func (a *AuthorizerSuite) TestToken() {
	id := "test-id"

	a.Run("valid", func() {
		token, err := a.authorizer.Token(id)
		a.Require().NoError(err)

		parsed, err := jwt.ParseWithClaims(token,
			&claims{},
			func(token *jwt.Token) (any, error) {
				return []byte(a.key), nil
			})
		a.Require().NoError(err)
		a.Require().True(parsed.Valid)

		c, ok := parsed.Claims.(*claims)
		a.Require().True(ok)
		a.Require().Equal(id, c.UserId)
	})
}

func (a *AuthorizerSuite) TestGetIdFromToken() {
	id := "test-id"

	a.Run("valid", func() {
		token, err := a.authorizer.Token(id)
		a.Require().NoError(err)

		parsedId, err := a.authorizer.GetIdFromToken(token)
		a.Require().NoError(err)
		a.Require().Equal(id, parsedId)
	})

	a.Run("wrong method", func() {
		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		a.Require().NoError(err)

		ttl := time.Now().Add(time.Minute)
		t := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(ttl),
			},
			UserId: id,
		})

		token, err := t.SignedString(privKey)
		a.Require().NoError(err)

		_, err = a.authorizer.GetIdFromToken(token)
		a.Require().ErrorIs(err, ErrUnexpectedMethod)
	})
}
