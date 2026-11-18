#CARA MEMBUAT REST API DENGAN GOLANG

1. Install Go dan jalankan "go mod init example.com atau url doamin/nama app atau folder"

2. Test dengan buat file main.go, dan buat function menjalankan "test" di terminal

3. Intall framework gin untuk pembuatan REST API, jalankan "go get -u github.com/gin-gonic/gin" di dalam folder backendnya

4. Test routing di dalam file main.go, import package gin-nya
   contoh:
   package main

import (
"net/http"
"github.com/gin-gonic/gin"
)

func main() {
router := gin.Default()

    // Route
    router.GET("/events", getEvents)
    router.GET("/test", getTest)

    // Jalankan server dan port
    router.Run(":8080")

}

// Function Handler
func getEvents(ctx \*gin.Context) {
ctx.JSON(http.StatusOK, gin.H {
"message": "Success get events",
})
}

func getTest(ctx \*gin.Context) {
ctx.JSON(http.StatusOK, gin.H {
"message": "Success get test",
})
}

5. Susnan folder

- Config untuk konek ke database
- Models untuk interaksi ke database
- Service untuk logic
- Controller untuk control routing/bisnis logic
- Router untuk patch url atau akses endpoint

6. Koneksi ke database postgresql (GORM) dan buat CRUD

- install GORM "go get -u gorm.io/gorm"
- install untuk postgresql "go get -u gorm.io/driver/postgres"
- install godotenv, buka github "https://github.com/joho/godotenv", lalu paste ke terminal "go get -u github.com/joho/godotenv"

7. Buat authentication dan protect route with jwt
- install package jwt-go "go get -u github.com/golang-jwt/jwt/v5"
- install bcrypt untuk hash password 1 arah "go get -u golang.org/x/crypto/bcrypt"

8. Installasi imagekit untuk keperluan upload image
- daftar akun imagekit
- pilih developer options
- cari cara pake go, nanti diarahkan ke github
- install diterminal "go get -u 'github.com/imagekit-developer/imagekit-go/v2@v2.14.0'"
- tari image_public_key, private_key, dan url ke env