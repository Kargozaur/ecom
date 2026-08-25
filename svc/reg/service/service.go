package service

import (
	"context"
	"errors"
	"pkg/token"
	userv1 "proto/out/user/v1"
	"reg/hasher"
	"reg/repo"
	"reg/util"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo      *repo.Repositories
	hasher    hasher.IHasher
	generator token.ITokenGenerator
	validator token.ITokenValidator
}

func New(db *pgxpool.Pool, hasher hasher.IHasher, generator token.ITokenGenerator, validator token.ITokenValidator) *Service {
	return &Service{repo: repo.NewRepositories(db), hasher: hasher, generator: generator, validator: validator}
}

func (s *Service) Register(ctx context.Context,
	params *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
	res := &userv1.RegisterUserResponse{}
	pwdHash, err := s.hasher.Hash(params.GetPassword())
	if err != nil {
		return nil, err
	}
	err = s.repo.WithinTx(ctx, func(ctx context.Context) error {
		_, err := s.repo.RegisterUser(ctx, params, pwdHash)
		if err != nil {
			return err
		}
		res.Response = "successfully registered"
		return nil
	})
	return res, nil
}

func (s *Service) Login(ctx context.Context,
	params *userv1.LoginUserRequest) (*userv1.LoginUserResponse, error) {
	var id string
	var email string
	var access string
	var refresh string
	iss := "user-service"
	e := errors.New("failed to validate credentials")
	err := s.repo.WithinTx(ctx, func(ctx context.Context) error {
		credentials, err := s.repo.LoginUser(ctx, params)
		if err != nil {
			return err
		}
		valid, err := s.hasher.CompareHashAndPassword(params.GetPassword(), credentials.PasswordHash)
		if err != nil {
			return err
		}
		if !valid {
			return e
		}
		id = credentials.ID.String()
		email = credentials.Email
		access, _ = s.generator.GenerateAccessToken(id, iss, email)
		refresh, _ = s.generator.GenerateRefreshToken(id, iss, email)
		tokenHash := hasher.HashToken(refresh)
		err = s.repo.CreateToken(ctx, id, tokenHash)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if util.AnyBlank(id, access, email, refresh) {
		return nil, e
	}
	return &userv1.LoginUserResponse{
		Access:  access,
		Refresh: refresh,
		Type:    "Bearer",
	}, nil
}

func (s *Service) FetchProfile(ctx context.Context,
	params *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	claims, err := s.validator.GetClaims(params.GetJwt(), token.Access)
	if err != nil {
		return nil, err
	}
	id, _ := uuid.Parse(claims.UserID)
	profile, err := s.repo.GetProfile(ctx, id)
	if err != nil {
		return nil, err
	}
	return &userv1.GetProfileResponse{
		Id:    profile.ID,
		Email: profile.Email,
		Name:  profile.Name,
		Date:  profile.CreatedAt.String(),
	}, nil
}

func (s *Service) Logout(ctx context.Context,
	params *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	tokenHash := hasher.HashToken(params.GetRefreshToken())
	err := s.repo.DeleteToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	return &userv1.LogoutResponse{
		Response: "Logged out successfully",
	}, nil
}
