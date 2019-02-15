package main

import(
	"database/sql"
	"os"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/Apinun/finalexam/customer"

	_ "github.com/lib/pq"
)

func createCustomersHandler(c *gin.Context)  {
	var item customer.Customer
	err := c.ShouldBindJSON(&item)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	row := db.QueryRow("INSERT INTO customer (name, email, status) VALUES ($1, $2, $3) RETURNING id", item.Name, item.Email, item.Status)
	err = row.Scan(&item.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "can't Scan row into variable" + err.Error()})
	}

	c.JSON(http.StatusCreated, item)

}


func getCustomersByIdHandler(c *gin.Context)  {
	id, _ := strconv.Atoi(c.Param("id"))

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer WHERE id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	row := stmt.QueryRow(id)

	t := customer.Customer{}
	err = row.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "data not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}


func getCustomersHandler(c *gin.Context)  {

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "can't query all customers" + err.Error()})
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "can't query all customers" + err.Error()})
		return
	}

	var customers = []customer.Customer{}

	for rows.Next(){
		t := customer.Customer{}
		err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "can't Scan row into variable" + err.Error()})
			return
		}

		customers = append(customers, t)
	}
	c.JSON(http.StatusOK, customers)
	// if len(customers) > 0 {
	// 	c.JSON(http.StatusOK, customers)
	// } else{
	// 	c.JSON(http.StatusUnauthorized, "Status code is 401 Unauthorized")
	// }	

}

func updateCustomersHandler(c *gin.Context)  {
	item := customer.Customer{}
	err := c.ShouldBindJSON(&item) // pass value type JSON to obj
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	stmt, err := db.Prepare("UPDATE customer SET name=$2 , email=$3 , status=$4 WHERE id=$1;")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	if _, err := stmt.Exec(id, item.Name, item.Email, item.Status); err!=nil{
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	item.ID= id
	c.JSON(http.StatusOK, item)
}

func deleteCustomersHandler(c *gin.Context)  {
	id, _ := strconv.Atoi(c.Param("id"))
	stmt, err := db.Prepare("DELETE FROM customer WHERE id =$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if _, err := stmt.Exec(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})

}

var db *sql.DB

func createTable()  {
	ctb := `
	CREATE TABLE IF NOT EXISTS customer (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`
	_, err := db.Exec(ctb)
	if err != nil {
		log.Fatal("can't create table", err)
	}
}

func loginMiddleware(c *gin.Context) {
	log.Println("Start Middleware")
	authKey := c.GetHeader("Authorization")
	if authKey != "token2019" {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
		c.Abort()
		return
	}
	c.Next()

}

func setUp() *gin.Engine  {
	r := gin.Default()
	r.Use(loginMiddleware)
	v1 := r.Group("")

	v1.POST("/customers" , createCustomersHandler)
	v1.GET("/customers/:id", getCustomersByIdHandler)
	v1.GET("/customers", getCustomersHandler)
	v1.PUT("/customers/:id", updateCustomersHandler)
	v1.DELETE("/customers/:id", deleteCustomersHandler)

	return r
}

func main()  {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("can't connect to database", err)
	}

	defer db.Close()
	createTable()

	r := setUp()
	r.Run(":2019")
}