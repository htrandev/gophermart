package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"

	"github.com/htrandev/gophermart/internal/application"
	"github.com/htrandev/gophermart/internal/domain"
	"github.com/htrandev/gophermart/pkg/luhn"
)

var errNilRequest = errors.New("request is nil")

type AuthorizationForm struct {
	User application.AuthorizationRequest
}

func (f *AuthorizationForm) BuildAndValidate(req *http.Request) error {
	if req == nil {
		return errNilRequest
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("register/BuildAndValidate: can't read body: %w", err)
	}
	defer req.Body.Close()

	var request application.AuthorizationRequest
	if err := easyjson.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("register/BuildAndValidate: can't unmarshal: %w", err)
	}

	if err := validation.ValidateStruct(&request,
		validation.Field(&request.Login, validation.Required),
		validation.Field(&request.Password,
			validation.Required,
			validation.Length(6, 72),
		),
	); err != nil {
		return fmt.Errorf("register/BuildAndValidate: invalid request: %w", err)
	}

	f.User = application.AuthorizationRequest{
		Login:    request.Login,
		Password: request.Password,
	}

	return nil
}

type AddOrderForm struct {
	Req domain.Order
}

func (f *AddOrderForm) BuildAndValidate(req *http.Request) error {
	if req == nil {
		return errNilRequest
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("addOrder/BuildAndValidate: can't read body: %w", err)
	}
	defer req.Body.Close()

	f.Req.Number = string(body)
	if err := validation.Validate(f.Req.Number, validation.Required); err != nil {
		return fmt.Errorf("addOrder/BuildAndValidate: invalid request: %w", err)
	}

	if !luhn.Check(f.Req.Number) {
		return domain.NewInvalidOrderError(f.Req.Number)
	}
	return nil
}

type WithdrawForm struct {
	Req domain.WithdrawRequest
}

func (f *WithdrawForm) BuildAndValidate(req *http.Request, uid uuid.UUID) error {
	if req == nil {
		return errNilRequest
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("withdraw/BuildAndValidate: can't read body: %w", err)
	}
	defer req.Body.Close()

	var request application.WithdrawRequest
	if err := easyjson.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("withdraw/BuildAndValidate: can't unmarshal body: %w", err)
	}

	if err := validation.ValidateStruct(&request,
		validation.Field(&request.Order, validation.Required),
		validation.Field(&request.Sum, validation.Required),
	); err != nil {
		return fmt.Errorf("validate request: %w", err)
	}

	if !luhn.Check(request.Order) {
		return domain.NewInvalidOrderError(request.Order)
	}

	f.Req = domain.WithdrawRequest{
		UserId: uid,
		Order:  request.Order,
		Sum:    request.Sum,
	}
	return nil
}
