package auth

import (
	authv1 "github.com/go-chat-devs/proto-auth-x-gateway/gen/go/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateRegister(req *authv1.RegisterRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	return nil
}


func ValidateLogin(req *authv1.LoginRequest) error {
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "empty email")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "empty password")
	}

	return nil
}

func ValidateSetupTOTP(req *authv1.SetupTOTPRequest) error {
	if len(req.Sessionkey) == 0{
		return status.Error(codes.InvalidArgument, "empty sessionkey")
	}
	return nil
}
func ValidateSetupTOTPValidate(req *authv1.SetupTOTPValidateRequest) error{
	if len(req.GetSetup2Fa()) == 0{
		return status.Error(codes.InvalidArgument,"empty 2fa")
	}
	if len(req.GetSessionkey()) == 0{
		return status.Error(codes.InvalidArgument,"empty sessionkey")
	}
	if len(req.GetCode()) == 0{
		return status.Error(codes.InvalidArgument,"empty code")
	}
	return nil
}

func ValidationTOTP(req *authv1.ValidateTOTPRequest) error{
	if len(req.Token2Fa) == 0{
		return status.Error(codes.InvalidArgument,"empty token2fa")
	}
	if len(req.GetCode()) == 0{
		return status.Error(codes.InvalidArgument,"empty code")
	}
	return nil
}


func ValidateLogout(req *authv1.LogoutRequest) error{
	if len(req.GetPassword()) == 0{
		return status.Error(codes.InvalidArgument,"empty password")
	}
	if len(req.GetSessionkey()) == 0{
		return status.Error(codes.InvalidArgument,"empty session key")
	}
	return nil
}