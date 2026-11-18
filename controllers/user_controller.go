package controllers

import (
	"net/http"
	"os"
	"time"

	"example.com/event-app/config"
	"example.com/event-app/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// untuk register
type AuthInputRegister struct {
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// untuk login
type AuthInputLogin struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

// untuk update profil
type AuthInputUpdate struct {
	Name  string `json:"name" binding:"omitempty"`
	Email string `json:"email" binding:"omitempty,email"`
}

// untuk ganti password
type AuthInputChangePassword struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"`
}

func RegisterUser(ctx *gin.Context) {
	// tampung ke variabel ambil dari struct nya
	var register AuthInputRegister

	// Validasi Register
	err := ctx.ShouldBindJSON(&register)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H {
			"message": err.Error(),
		})

		return
	}

	// Hash Password (Enkripsi 1 arah)
	hashedPassword, errHash := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)

	if errHash != nil {
		ctx.JSON(http.StatusBadRequest, gin.H {
			"message": "Failed to hash password",
		})

		return
	}

	// Simpan ke Database
	user := models.User {
		Name: register.Name,
		Email: register.Email,
		Password: string(hashedPassword),
	}

	createUser := config.DB.Create(&user)

	if createUser.Error != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to create user",
			"error":   createUser.Error.Error(), // sementara, untuk debugging
		})

		return
	}

	// Response Success
	ctx.JSON(http.StatusCreated, gin.H {
		"message": "User created successfully",
		"user": gin.H {
			"id": user.ID,
			"name": user.Name,
			"email": user.Email,
			"events": user.Events,
		},
	})
}

func LoginUser(ctx *gin.Context) {
	// tampung ke variabel ambil dari struct nya
	var login AuthInputLogin

	// Validasi Login
	err := ctx.ShouldBindJSON(&login)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H {
			"message": err.Error(),
		})

		return
	}

	// cek apakah email ada di database
	var user models.User

	userData := config.DB.Where("email = ?", login.Email).First(&user).Error

	// jika user belum ada di database
	if userData != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H {
			"message": "Email is wrong",
		})

		return
	}

	// cek password
	errMatchPassword := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))

	if errMatchPassword != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H {
			"message": "Password is wrong",
		})

		return
	}

	// Buat token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,   // sebelumnya "sub"
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	// generate token
	tokenString, errGenerateToken := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if errGenerateToken != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H {
			"message": "Failed to generate token",
		})

		return
	}

	// Response Success
	ctx.JSON(http.StatusOK, gin.H {
		"message": "Login successfully",
		"user": gin.H {
			"id": user.ID,
			"name": user.Name,
			"email": user.Email,
			"events": user.Events,
		},
		"token": tokenString,
	})
}

func GetCurrentUser(ctx *gin.Context) {
	// ambil user id dari token
	userId, exists := ctx.Get("userID")

	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized: please log in first",
		})
		return
	}

	// ambil user dari database
	var user models.User

	if err := config.DB.Select("id", "name", "email").First(&user, userId).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	// response success
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Current user retrieved successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func UpdateCurrentUser(ctx *gin.Context) {
	userId, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: please log in first"})
		return
	}

	// validasi data pakai struct input (bukan model User)
	var input AuthInputUpdate
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// ambil user dari database
	var user models.User
	if err := config.DB.First(&user, userId).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// update hanya name & email (field kosong diabaikan)
	if err := config.DB.Model(&user).Updates(models.User{
		Name:  input.Name,
		Email: input.Email,
	}).Error; err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"message": "Failed to update user, email may already be in use"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Current user updated successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func ChangePassword(ctx *gin.Context) {
	userId, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: please log in first"})
		return
	}

	var input AuthInputChangePassword
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userId).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// cek password lama
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Old password is wrong"})
		return
	}

	// hash password baru
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
		return
	}

	// simpan password baru
	if err := config.DB.Model(&user).Update("password", string(hashed)).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}