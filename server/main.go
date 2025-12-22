package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_"github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)

type Quote struct {
	Author string
	Quote  string
	ID     int
	Trunc  string
	isLong bool
	Tags   []string
}

	// Disable Console Color
	// gin.DisableConsoleColor()
	items := []Quote{
		{"C.S. Lewis", "Something or another about being humble but it's actually Rick Warren this quote is really long just making this extra long to showcase a longer quote for the sake of an insane length", 1,
			"Something or another about being humble but it's actually Rick Warren", true, []string{"inpsiring", "faith"}},
		{"JRR Tolkien", "Lots of lunches and such", 2, "Lots of lunches and such", false, []string{"fantasy", "wow"}},
		{"Gene Wolfe", "My main character is so interesting, mostly because he's insane", 3, "My main character is so interesting, mostly because he's insane", false, []string{"soneat", "wow"}},
		{"Dean Koontz", "Wow tire salesman", 4, "Wow tire salesman", false, []string{"tires", "wow"}},
	}
	r := gin.Default()
	r.LoadHTMLFiles("../client/index.html")
	r.Static("/css", "../client/css")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{"items": items})
	})

	// Get user value
	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		value, ok := db[user]
		if ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": value})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	// Authorized group (uses gin.BasicAuth() middleware)
	// Same than:
	// authorized := r.Group("/")
	// authorized.Use(gin.BasicAuth(gin.Credentials{
	//	  "foo":  "bar",
	//	  "manu": "123",
	//}))
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"foo":  "bar", // user:foo password:bar
		"manu": "123", // user:manu password:123
	}))

	/* example curl for /admin with basicauth header
	   Zm9vOmJhcg== is base64("foo:bar")

		curl -X POST \
	  	http://localhost:8080/admin \
	  	-H 'authorization: Basic Zm9vOmJhcg==' \
	  	-H 'content-type: application/json' \
	  	-d '{"value":"bar"}'
	*/
	authorized.POST("admin", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)

		// Parse JSON
		var json struct {
			Value string `json:"value" binding:"required"`
		}

		if c.Bind(&json) == nil {
			db[user] = json.Value
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
