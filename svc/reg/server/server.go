package server

import (
	"context"
	"errors"
	"pkg/token"
	userv1 "proto/out/user/v1"
	"reg/hasher"
	"reg/service"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	userv1.UnimplementedUserServiceServer
	service *service.Service
}

func NewGRPCServer(pool *pgxpool.Pool) *GRPCServer {
	cfg, err := token.NewTokenConfig()
	if err != nil {
		return nil
	}
	hasher := hasher.NewArgon2Hasher()
	tokenGenerator := token.NewTokenGenerator(cfg, "user-service")
	tokenValidator := token.NewTokenValidator(cfg)
	service := service.New(pool, hasher, tokenGenerator, tokenValidator)
	return &GRPCServer{
		service: service,
	}
}

func (s *GRPCServer) RegisterUser(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	r, err := s.service.Register(ctx, req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return r, nil
}

func (s *GRPCServer) LoginUser(ctx context.Context, req *userv1.LoginUserRequest) (*userv1.LoginUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	r, err := s.service.Login(ctx, req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return r, nil
}

func (s *GRPCServer) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	r, err := s.service.FetchProfile(ctx, req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return r, nil
}

func (s *GRPCServer) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	r, err := s.service.Logout(ctx, req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return r, nil
}
