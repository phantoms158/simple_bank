package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/phantoms158/simple_bank/db/sqlc"
	"github.com/phantoms158/simple_bank/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type createUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errResponse(err))
		return
	}

	hashPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashPassword,
		Email:          req.Email,
		FullName:       req.FullName,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code.Name() {
			case "unique_violation":
				{
					ctx.JSON(http.StatusForbidden, errResponse(err))
					return
				}
			}
		}
		ctx.JSON(http.StatusInternalServerError, errResponse(err))
		return
	}
	rsp := createUserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}
