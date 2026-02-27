package auth

import (
	"context"
	"errors"

	authv1 "github.com/go-chat-devs/proto-auth-x-gateway/gen/go/auth"
	"github.com/go-chat-devs/service-auth/internal/token"
	"github.com/pquerna/otp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Register(
		ctx context.Context,
		email string,
		password string,
	) (err error)
	Login(
		ctx context.Context,
		email string,
		password string,
	) (token token.Token, err error)

	SetupTOTP(
		ctx context.Context,
		sessionkey token.Token,
	) (res struct {
		Key   *otp.Key
		Token token.Token
	}, err error)
	SetupTOTPValidate(
		ctx context.Context,
		sessionkey token.Token,
		setup2fa token.Token,
		code string,
	) (err error)
	ValidateTOTP(
		ctx context.Context,
		token2fa token.Token,
		code string,
	) (sessionkey token.Token, err error)
	Logout(
		ctx context.Context,
		sessionkey token.Token,
		password string,
	) (err error)
}

type ServerApi struct {
	authv1.UnimplementedAuthServer
	auth Auth
}

func RegisterServerApi(gRPC *grpc.Server, auth Auth) {
	authv1.RegisterAuthServer(gRPC, &ServerApi{auth: auth})

}

func (s *ServerApi) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := ValidateRegister(req); err != nil {
		return nil, err
	}
	err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, errors.New("user already exists")) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &authv1.RegisterResponse{}, nil

}

func (s *ServerApi) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if err := ValidateLogin(req); err != nil {
		return nil, err
	}
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, errors.New("Invalid credentials")) {
			return nil, status.Error(codes.InvalidArgument, "invalid arguments")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.LoginResponse{
		Token: token[:],
	}, nil
}

func (s *ServerApi) SetupTOTP(ctx context.Context, req *authv1.SetupTOTPRequest) (*authv1.SetupTOTPResponse, error) {
	if err := ValidateSetupTOTP(req); err != nil {
		return nil, err
	}
	res, err := s.auth.SetupTOTP(ctx, token.Token(req.GetSessionkey()))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	grpcURL := &authv1.Url{
		Url: res.Key.URL(),
	}
	key := &authv1.Key{
		Orig: res.Key.String(),
		Url:  grpcURL,
	}
	return &authv1.SetupTOTPResponse{
		Res: &authv1.Res{
			Key:   key,
			Token: res.Token[:],
		},
	}, nil
}

func (s *ServerApi) SetuptTOTPValidate(ctx context.Context, req *authv1.SetupTOTPValidateRequest) (*authv1.SetupTOTPValidateResponse, error) {
	if err := ValidateSetupTOTPValidate(req); err != nil {
		return nil, err
	}
	err := s.auth.SetupTOTPValidate(ctx, token.Token(req.GetSessionkey()), token.Token(req.GetSetup2Fa()), req.GetCode())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal")
	}
	return &authv1.SetupTOTPValidateResponse{}, nil
}

func (s *ServerApi) ValidateTOTP(ctx context.Context, req *authv1.ValidateTOTPRequest) (*authv1.ValidateTOTPResponse, error) {
	if err := ValidationTOTP(req); err != nil {
		return nil, err
	}
	sessionkey, err := s.auth.ValidateTOTP(ctx, token.Token(req.GetToken2Fa()), req.GetCode())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &authv1.ValidateTOTPResponse{
		Sessionkey: sessionkey[:],
	}, nil
}

func (s *ServerApi) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := ValidateLogout(req); err != nil {
		return nil, err
	}
	err := s.auth.Logout(ctx, token.Token(req.GetSessionkey()), req.GetPassword())
	if err != nil {
		if errors.Is(err, errors.New("invalid credentials")) {
			return nil, status.Error(codes.InvalidArgument, "invalid arguments")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &authv1.LogoutResponse{}, nil
}
