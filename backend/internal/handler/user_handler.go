package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"backend/internal/model"
	"backend/internal/service"
	"backend/pkg/response"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Address  string `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.service.CreateUser(req.Username, req.Email, req.Address)
	if err != nil {
		response.Error(c, 1001, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	user, err := h.service.GetUser(uint(id))
	if err != nil {
		response.Error(c, 1002, err.Error())
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.ListUsers()
	if err != nil {
		response.Error(c, 1003, err.Error())
		return
	}

	response.Success(c, users)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Address  string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Address:  req.Address,
	}
	user.ID = uint(id)

	if err := h.service.UpdateUser(user); err != nil {
		response.Error(c, 1004, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.service.DeleteUser(uint(id)); err != nil {
		response.Error(c, 1005, err.Error())
		return
	}

	response.Success(c, nil)
}
