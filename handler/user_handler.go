package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"web2/config"

	"github.com/allegro/bigcache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

var (
	GlobalCache *bigcache.BigCache
	DB          *gorm.DB
)

func init() {
	//open a db connection
	var err error
	DB, err = gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=postgres password=123456")
	if err != nil {
		panic("failed to connect database")
	}
	//Migrate the schema
	DB.AutoMigrate(&userModel{})
}

// createTodo add a new todo
func CreateUser(c *gin.Context) {
	completed, _ := strconv.Atoi(c.PostForm("completed"))
	user := userModel{UserName: c.PostForm("userName"), Password: c.PostForm("password"), Role: c.PostForm("role"), Completed: completed}
	DB.Save(&user)
	c.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "message": "User created successfully!", "resourceId": user.ID})
}

// getAllUser
func GetAllUser(c *gin.Context) {
	var users []userModel
	var _users []transformedUser
	DB.Find(&users)
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
func GetUser(c *gin.Context) {
	var user userModel
	userID := c.Param("id")
	DB.First(&user, userID)
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
func ModifyUser(c *gin.Context) {
	var user userModel
	userID := c.Param("id")
	DB.First(&user, userID)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No user found!"})
		return
	}
	DB.Model(&user).Update("userName", c.PostForm("userName"))
	DB.Model(&user).Update("password", c.PostForm("password"))
	DB.Model(&user).Update("role", c.PostForm("role"))
	completed, _ := strconv.Atoi(c.PostForm("completed"))
	DB.Model(&user).Update("completed", completed)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "User updated successfully!"})
}

// delete an user
func DeleteUser(c *gin.Context) {
	var user userModel
	userID := c.Param("id")
	DB.First(&user, userID)
	if user.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "No user found!"})
		return
	}
	DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "message": "User deleted successfully!"})
}

//Login
func Login(c *gin.Context) {
	var user userModel
	username, password := c.PostForm("username"), c.PostForm("password")
	// If user has logged in, force him to log out firstly
	for iter := GlobalCache.Iterator(); iter.SetNext(); {
		info, err := iter.Value()
		if err != nil {
			continue
		}
		if string(info.Value()) == username {
			GlobalCache.Delete(info.Key())
			log.Printf("forced %s to log out\n", username)
			break
		}
	}
	DB.Where("user_name = ? AND password = ?", username, password).First(&user)
	fmt.Println(user)
	if user.Role == "admin" {
		log.Println(fmt.Sprintf("%s has logged in.", username))
	} else if user.Role == "user" {
		log.Println(fmt.Sprintf("%s has logged in.", username))
	} else {
		c.JSON(200, config.RestResponse{Message: "no such account"})
		return
	}

	// // Apparently we don't do this in real world :)
	// if username == "alice" && password == "111" {
	// 	log.Println(fmt.Sprintf("%s has logged in.", username))
	// } else if username == "bob" && password == "123" {
	// 	log.Println(fmt.Sprintf("%s has logged in.", username))
	// } else {
	// 	c.JSON(200, config.RestResponse{Message: "no such account"})
	// 	return
	// }

	// Generate random session id
	u, err := uuid.NewRandom()
	if err != nil {
		log.Println(fmt.Errorf("failed to generate UUID: %w", err))
	}
	sessionId := fmt.Sprintf("%s-%s", u.String(), username)
	// Store current subject in cache
	err = GlobalCache.Set(sessionId, []byte(username))
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to store current subject in cache: %w", err))
		return
	}
	// Send session id back to client in cookie
	c.SetCookie("current_subject", sessionId, 30*60, "/api/v1", "", 1, false, true)
	c.JSON(200, config.RestResponse{Code: 1, Message: username + " logged in successfully"})
}
