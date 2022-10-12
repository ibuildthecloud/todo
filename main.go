package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db      *gorm.DB
	migrate = flag.Bool("migrate", false, "Run DB migration and exit")
)

type Item struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Title     string         `json:"title"`
	Completed bool           `json:"completed"`
}

func deleteItem(key string) error {
	i, err := getItem(key)
	if err != nil {
		return err
	}
	return db.Delete(i).Error
}

func updateItem(item *Item, key string) (*Item, error) {
	id, err := strconv.Atoi(key)
	if err != nil {
		return nil, err
	}
	item.ID = uint(id)
	return item, db.Save(item).Error
}

func getItem(key string) (item *Item, _ error) {
	return item, db.First(&item, key).Error
}

func create(i *Item) error {
	return db.Create(i).Error
}

func getAll() (item []Item, _ error) {
	return item, db.Find(&item).Error
}

func initDB() error {
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbParams := os.Getenv("DB_PARAMS")

	if dbHost == "" {
		localDB, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
		if err != nil {
			return err
		}
		db = localDB
	} else {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", dbUser, "****", dbHost, dbPort, dbName, dbParams)
		log.Println("Connecting to DB " + dsn)
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", dbUser, dbPass, dbHost, dbPort, dbName, dbParams)
		localDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return err
		}
		db = localDB
	}

	if *migrate || os.Getenv("GIN_MODE") != "release" {
		log.Println("Running DB migration")
		if err := db.AutoMigrate(&Item{}); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	if err := initDB(); err != nil {
		log.Fatal(err)
	}

	if *migrate {
		return
	}

	r := gin.Default()

	r.PUT("/todos/:id", func(c *gin.Context) {
		item := &Item{}
		if err := c.BindJSON(item); err != nil {
			c.AbortWithStatusJSON(500, err)
			return
		}

		i, err := updateItem(item, c.Param("id"))
		if err == nil {
			c.JSON(200, i)
		} else {
			c.Status(404)
		}
	})

	r.GET("/todos/:id", func(c *gin.Context) {
		i, err := getItem(c.Param("id"))
		if err == nil {
			c.JSON(200, i)
		} else {
			c.Status(404)
		}
	})

	r.DELETE("/todos/:id", func(c *gin.Context) {
		err := deleteItem(c.Param("id"))
		if err == nil {
			c.Status(200)
		} else {
			c.AbortWithError(500, err)
		}
	})

	r.POST("/todos", func(c *gin.Context) {
		item := &Item{}
		if err := c.BindJSON(item); err != nil {
			c.AbortWithStatusJSON(500, err)
			return
		}

		if err := create(item); err != nil {
			c.AbortWithStatusJSON(500, err)
			return
		}

		c.JSON(200, item)
	})

	r.GET("/todos", func(c *gin.Context) {
		i, err := getAll()
		if err == nil {
			c.JSON(200, i)
		} else {
			c.AbortWithError(500, err)
		}
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
