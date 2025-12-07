package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
	"github.com/htrandev/gophermart/pkg/strutil"
)

type AuthorizationSuite struct {
	ServiceSuite
}

func (s *AuthorizationSuite) SetupSuite() {
	s.ServiceSuite.SetupSuite()
}

func (s *AuthorizationSuite) TearDownSuite() {
	s.service.Close()
}

func TestAuthorization(t *testing.T) {
	suite.Run(t, new(AuthorizationSuite))
}

func (s *AuthorizationSuite) TestRegister() {
	ctx := context.Background()

	id := "test_id"
	token := "token"
	hash := "hash"

	u := application.AuthorizationRequest{
		Login:    "test",
		Password: "password",
	}

	s.Run("valid", func() {
		s.authorizer.EXPECT().HashPassword(u.Password).Return(hash, nil).Times(1)
		s.repository.EXPECT().Register(gomock.Any(), domain.User{
			Login:        u.Login,
			HashPassword: hash,
		}).Return(id, nil).Times(1)
		s.authorizer.EXPECT().Token(id).Return(token, nil).Times(1)

		t, err := s.service.Register(ctx, u)
		s.Require().NoError(err)
		s.Require().Equal(token, t)
	})

	s.Run("long password", func() {
		longPass := strutil.Random(100)

		s.authorizer.EXPECT().HashPassword(longPass).Return("", bcrypt.ErrPasswordTooLong).Times(1)

		_, err := s.service.Register(ctx, application.AuthorizationRequest{
			Password: longPass,
		})
		s.Require().ErrorIs(err, bcrypt.ErrPasswordTooLong)
	})

	s.Run("repo err", func() {
		s.authorizer.EXPECT().HashPassword(u.Password).Return(hash, nil).Times(1)
		s.repository.EXPECT().Register(gomock.Any(), domain.User{
			Login:        u.Login,
			HashPassword: hash,
		}).Return("", errTest).Times(1)

		_, err := s.service.Register(ctx, u)
		s.Require().ErrorIs(err, errTest)
	})

	s.Run("token error", func() {
		s.authorizer.EXPECT().HashPassword(u.Password).Return(hash, nil).Times(1)
		s.repository.EXPECT().Register(gomock.Any(), domain.User{
			Login:        u.Login,
			HashPassword: hash,
		}).Return(id, nil).Times(1)
		s.authorizer.EXPECT().Token(id).Return("", errTest).Times(1)

		_, err := s.service.Register(ctx, u)
		s.Require().ErrorIs(err, errTest)
	})
}

func (s *AuthorizationSuite) TestLogin() {
	ctx := context.Background()

	token := "token"

	u := application.AuthorizationRequest{
		Login:    "test",
		Password: "password",
	}
	dbUser := domain.User{
		Id:           uuid.New(),
		Login:        "test",
		HashPassword: "hash",
	}

	s.Run("valid", func() {
		s.repository.EXPECT().Login(gomock.Any(), u.Login).Return(dbUser, nil)
		s.authorizer.EXPECT().ValidatePassword(dbUser.HashPassword, u.Password).Return(true).Times(1)
		s.authorizer.EXPECT().Token(dbUser.Id.String()).Return(token, nil).Times(1)

		t, err := s.service.Login(ctx, u)
		s.Require().NoError(err)
		s.Require().Equal(token, t)
	})

	s.Run("repository error", func() {
		s.repository.EXPECT().Login(gomock.Any(), u.Login).Return(domain.User{}, errTest)

		_, err := s.service.Login(ctx, u)
		s.Require().ErrorIs(err, errTest)
	})

	s.Run("validate password error", func() {
		s.repository.EXPECT().Login(gomock.Any(), u.Login).Return(domain.User{
			HashPassword: "wrong password",
		}, nil)
		s.authorizer.EXPECT().ValidatePassword("wrong password", u.Password).Return(false).Times(1)

		_, err := s.service.Login(ctx, u)
		s.Require().ErrorIs(err, domain.ErrIncorrectPassword)
	})

	s.Run("token error", func() {
		s.repository.EXPECT().Login(gomock.Any(), u.Login).Return(dbUser, nil)
		s.authorizer.EXPECT().ValidatePassword(dbUser.HashPassword, u.Password).Return(true).Times(1)
		s.authorizer.EXPECT().Token(dbUser.Id.String()).Return("", errTest).Times(1)

		_, err := s.service.Login(ctx, u)
		s.Require().ErrorIs(err, errTest)
	})
}
