package gapi

import (
	"context"
	"database/sql"

	"github.com/phantoms158/simple_bank/pb"
	"github.com/phantoms158/simple_bank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "not found user")
		}
		return nil, status.Errorf(codes.Internal, "falied to find user error")
	}

	err = util.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "incorrect password")
	}

	accessToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}

	rsp := &pb.LoginUserResponse{
		User:        convertUser(user),
		AccessToken: accessToken,
	}
	return rsp, nil
}
