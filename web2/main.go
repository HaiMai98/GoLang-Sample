package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type (
	userModel struct {
		gorm.Model
		UserName  string `json:"userName"`
		Password  string `json:"password"`
		Role      string `json:"role"`
		Completed int    `json:"completed"`
	}

	transformedUser struct {
		ID        string `json:"id"`
		UserName  string `json:"userName"`
		Password  string `json:"password"`
		Role      string `json:"role"`
		Completed bool   `json:"completed"`
	}
)

var db *gorm.DB

func init() {
	//open a db connection
	var err error
	db, err = gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=postgres password=123456")
	if err != nil {
		panic("failed to connect database")
	}
	//Migrate the schema
	db.AutoMigrate(&userModel{})
}

// createTodo add a new todo
func createUser(c *gin.Context) {
	completed, _ := strconv.Atoi(c.PostForm("completed"))
	user := userModel{UserName: c.PostForm("userName"), Password: c.PostForm("password"), Role: c.PostForm("role"), Completed: completed}
	db.Save(&user)
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "User created successfully!", "resourceId": user.ID})
}

// getAllUser
func getAllUser(c *gin.Context) {
	var users []userModel
	var _users []transformedUser
	db.Find(&users)
	if len(users) <= 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No User found!"})
		return
	}
	//transforms the users for building a good response
	for _, item := range users {
		completed := false
		if item.Completed == 1 {
			completed = true
		} else {
			completed = false
		}
		_users = append(_users, transformedUser{UserName: item.UserName, Password: item.Password, Role: item.Role, Completed: completed})
	}
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _users})
}

// get an user
func getUser(c *gin.Context) {
	var user userModel
	userID := c.Param("id")
	db.First(&user, userID)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No todo found!"})
		return
	}
	completed := false
	if user.Completed == 1 {
		completed = true
	} else {
		completed = false
	}
	_user := transformedUser{UserName: user.UserName, Password: user.Password, Role: user.Role, Completed: completed}
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": _user})
}

// modify User
func modifyUser(c *gin.Context) {
	var user userModel
	userID := c.Param("id")
	db.First(&user, userID)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No user found!"})
		return
	}
	db.Model(&user).Update("userName", c.PostForm("userName"))
	db.Model(&user).Update("password", c.PostForm("password"))
	db.Model(&user).Update("role", c.PostForm("role"))
	completed, _ := strconv.Atoi(c.PostForm("completed"))
	db.Model(&user).Update("completed", completed)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "User updated successfully!"})
}

// delete an user
func deleteUser(c *gin.Context) {
	var user userModel
	userID := c.Param("id")
	db.First(&user, userID)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No user found!"})
		return
	}
	db.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "User deleted successfully!"})
}

func main() {
	router := gin.Default()
	v1 := router.Group("api/v1/web")
	{
		v1.POST("/", createUser)
		v1.GET("/", getAllUser)
		v1.GET("/:id", getUser)
		v1.PUT("/:id", modifyUser)
		v1.DELETE("/:id", deleteUser)
	}
	router.Run()
}
