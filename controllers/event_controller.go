package controllers

import (
	"context"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"example.com/event-app/config"
	"example.com/event-app/models"
	"github.com/gin-gonic/gin"
	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"gorm.io/gorm"
)

func InitImageKit() *imagekit.Client {
	client := imagekit.NewClient(
		option.WithPrivateKey(os.Getenv("IMAGEKIT_PRIVATE_KEY")),
	)
	return &client
}

func GetEvents(ctx *gin.Context) {
	// ambil semua data dari database dengan awalan array kosong
	var events []models.Event

	// fitur seacrh atau pagination

	// 1. insialisasi dasar query di gorm
	query := config.DB.Model(&models.Event{})

	// 2. tangkap fungsi filter by query
	seacrh := ctx.Query("search")

	// validasi
	if seacrh != "" {
		query = query.Where("name_event ILIKE ? OR location ILIKE ?", "%"+seacrh+"%", "%"+seacrh+"%")
	}

	// 3. hitung total data sebelum di limit (pagination)
	var totalRows int64

	query.Count(&totalRows)

	// 4. tangkap page dan limit
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "6")

	// konversi jadi interger
	page, errPage := strconv.Atoi(pageStr)

	if errPage != nil || page < 1 {
		page = 1
	}

	limit, errLimit := strconv.Atoi(limitStr)

	if errLimit != nil || limit < 1 {
		limit = 6
	}

	// 5. hitung offset
	offset := (page - 1) * limit

	// 6. hitung data per page
	totalPage := int(math.Ceil(float64(totalRows) / float64(limit)))

	// 7. eksekusi semua fitur yang dibuat diatas
	if err := query.Preload("User", func (db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H {
			"message": "Failed to get events",
		})

		return
	}

	// // panggil fungsi getAll untuk mengambil semua data
	// config.DB.Find(&events)

	// kirim response
	ctx.JSON(http.StatusOK, gin.H {
		"message": "Get all events successfully",
		"events": events,
		"meta": gin.H {
			"page": page,
			"limit": limit,
			"totalRows": totalRows,
			"totalPage": totalPage,
		},
	})
}

func GetEventById(ctx *gin.Context) {
	// ambil data berdasarkan id, ambil struct event bukan array
	var event models.Event

	// ambil parameter id, sesuaikan dengan route kalo "id" tulis "id", kalo "event_id" tulis "event_id"
	paramId := ctx.Param("id")

	// kalo datanya tidak ditemukan
	var eventData = config.DB.Preload("User", func (db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Preload("Bookings").Preload("Bookings.User", func (db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).First(&event, paramId).Error

	// jika datanya tidak ditemukan
	if eventData != nil {
		ctx.JSON(http.StatusNotFound, gin.H {
			"message": "Event not found",
		})
		return
	}

	// panggil fungsi get berdasarkan id
	config.DB.First(&event, paramId)

	// kirim response
	ctx.JSON(http.StatusOK, gin.H {
		"message": "Get event by id successfully",
		"event": event,
	})
}

func GetEventsByUser (ctx *gin.Context) {
	var events []models.Event

	userId, _ := ctx.Get("userID")

	err := config.DB.Preload("User", func (db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Where("user_id = ?", userId).Find(&events).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H {
			"message": "Failed to get events by user",
		})
		return
	}

	// response berhasil
	ctx.JSON(http.StatusOK, gin.H {
		"message": "Get events by user successfully",
		"events": events,
	})
}

func CreateEvent(ctx *gin.Context) {
	// ambil data user yang sudah login
	userId, _ := ctx.Get("userID")

	// menerima file form-data
	file, header, err := ctx.Request.FormFile("image")

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Image is required",
		})
		return
	}

	defer file.Close()

	// upload file ke imagekit
	fileName := header.Filename

	ik := InitImageKit()

	uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
		File: file,
		FileName: fileName,
	})

	if errUpload != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload image",
		})
		return
	}

	// parsing tanggal jangan diabaikan
	parsedTime, err := time.Parse(time.RFC3339, ctx.PostForm("date_time"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "date_time must be RFC3339, e.g. 2026-11-27T16:16:20Z"})
		return
	}

	// Simpan kedatabase
	event := models.Event{
		NameEvent: ctx.PostForm("name_event"),
		Description: ctx.PostForm("description"),
		Location: ctx.PostForm("location"),
		Image: uploadRes.URL,
		ImageId: uploadRes.FileID,
		UserId: userId.(int),
		DateTime: parsedTime,
	}

	// tanpa upload image
	// // ambil data dari request atau body
	// var event models.Event

	// // validasi data
	// err := ctx.ShouldBindJSON(&event)

	// // jika error kirim error
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	// 	return
	// }

	// // ambil id user
	// event.UserId = userId.(int)

	// panggil fungsi create untuk menyimpan data
	config.DB.Create(&event)

	// kirim response
	ctx.JSON(http.StatusCreated, gin.H {
		"message": "Event created successfully",
		"event": event,
	})
}

func UpdateEvent(ctx *gin.Context) {
	// ambil data user yang sudah login
	userId, _ := ctx.Get("userID")

	// ambil data dari request atau body
	var event models.Event

	// ambil parameter id, sesuaikan dengan route kalo "id" tulis "id", kalo "event_id" tulis "event_id"
	paramId := ctx.Param("id")

	// kalo datanya tidak ditemukan
	var eventData = config.DB.First(&event, paramId).Error

	// jika datanya tidak ditemukan
	if eventData != nil {
		ctx.JSON(http.StatusNotFound, gin.H {
			"message": "Event not found",
		})
		return
	}

	// cek apakah user yang update sama dengan user yang login
	if event.UserId != userId.(int) {
		ctx.JSON(http.StatusForbidden, gin.H {
			"message": "You are not authorized to update this event",
		})
		return
	}

	// update image
	file, header, err := ctx.Request.FormFile("image")

	if err == nil {
		defer file.Close();

		// buat variabel
		ik := InitImageKit()

		// upload file gambar new
		fileName := header.Filename

		uploadRes, errUpload := ik.Files.Upload(context.Background(), imagekit.FileUploadParams{
			File: file,
			FileName: fileName,
		})

		if errUpload == nil {
			// Hapus gambar lama
			if event.ImageId != "" {
				ik.Files.Delete(context.Background(), event.ImageId)
			}

			// upload field gambar
			event.Image = uploadRes.URL
			event.ImageId = uploadRes.FileID
		}
	}

	if name := ctx.PostForm("name"); name != "" {
		event.NameEvent = name
	}

	if description := ctx.PostForm("description"); description != "" {
		event.Description = description
	}

	if location := ctx.PostForm("location"); location != "" {
		event.Location = location
	}

	if dateTime := ctx.PostForm("date_time"); dateTime != "" {
		parsedTime, err := time.Parse(time.RFC3339, dateTime)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "date_time must be RFC3339, e.g. 2026-11-27T16:16:20Z"})
			return
		}
		event.DateTime = parsedTime
	}

	// tanpa upload image
	// // validasi data
	// err := ctx.ShouldBindJSON(&event)

	// // jika error kirim error
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	// 	return
	// }

	// panggil fungsi update untuk mengupdate data
	config.DB.Save(&event)

	// kirim response
	ctx.JSON(http.StatusOK, gin.H {
		"message": "Event updated successfully",
		"event": event,
	})
}

func DeleteEvent(ctx *gin.Context) {
	// ambil data user yang sudah login
	userId, _ := ctx.Get("userID")

	// ambil data berdasarkan id, ambil struct event bukan array
	var event models.Event

	// ambil parameter id, sesuaikan dengan route kalo "id" tulis "id", kalo "event_id" tulis "event_id"
	paramId := ctx.Param("id")

	// kalo datanya tidak ditemukan
	var eventData = config.DB.First(&event, paramId).Error

	// jika datanya tidak ditemukan
	if eventData != nil {
		ctx.JSON(http.StatusNotFound, gin.H {
			"message": "Event not found",
		})
		return
	}

	// cek apakah user yang delete sama dengan user yang login
	if event.UserId != userId.(int) {
		ctx.JSON(http.StatusForbidden, gin.H {
			"message": "You are not authorized to delete this event",
		})
		return
	}

	// hapus image
	if event.ImageId != "" {
		ik := InitImageKit()
		ik.Files.Delete(context.Background(), event.ImageId)
	}

	// panggil fungsi delete untuk menghapus data, kenapa pake unscoped? karena dia soft delete di supabase
	config.DB.Unscoped().Delete(&event, paramId)

	// kirim response
	ctx.JSON(http.StatusOK, gin.H {
		"message": "Event deleted successfully",
	})
}