package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"github.com/ttacon/libphonenumber"
)
type getUserRequest struct {
	Name string `json:"full_name"`
	Email string `json:"email"`
	Contacts []string `json:"phone_numbers"`
}

type postUserRequest struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Contacts []string `json:"phone_numbers"`
}

type Contacts struct {
	Number string
	UserId int
}


func getUsers(ctx *gin.Context) {
	db := SetupDB()
	rows, err := db.Query("SELECT  CONCAT(users.first_name , ' ' ,users.last_name) as Name, users.email as Email, ARRAY_AGG(contacts.number) as Contacts FROM users JOIN contacts on users.id=contacts.user_id GROUP BY Name, Email")
	checkErr(err, "error while querying users")
	defer rows.Close()
	users := []getUserRequest{}

	for rows.Next(){
		checkErr(rows.Err(),"Error while getting db rows")
		u := getUserRequest{}
		err := rows.Scan(&u.Name, &u.Email, pq.Array(&u.Contacts))
		checkErr(err,"Error while scanning user data")
		users = append(users, u)
	}
	ctx.IndentedJSON(http.StatusOK, users)

}

func validateRequest( user postUserRequest) (bool,error){
	fmt.Println(user)
	if user.Email == "" || ! strings.Contains(user.Email,"@"){
		return false, errors.New("Incorrect or missing email id")
	}
	if user.FirstName == "" {
		return false, errors.New("Missing first name")
	}
	fmt.Println("failed pho", len(user.Contacts))
	if len(user.Contacts) == 0 {
		return false, errors.New("Empty phone_numbers list")
	}else if len(user.Contacts) > 0  {
		for _, contact:= range user.Contacts{
			phoneNumber, contacterr := libphonenumber.Parse(contact,"AU")
			if contacterr!=nil{
				return false, contacterr
			}
			if !libphonenumber.IsValidNumber(phoneNumber) {
				return false, errors.New("Invalid Mobile Number "+contact)
			}
		} 
	}
	
	
	return true, nil
}



func addUsers(ctx *gin.Context){
	db := SetupDB()
	tx, txerr := db.Begin()
	if txerr!=nil{
		log.Panicln(txerr)
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":txerr.Error()})
	}

	user := postUserRequest{}
	ctx.BindJSON(&user)
	valid, validationerr:=validateRequest(user)
	fmt.Println(valid, " ", validationerr)
	if validationerr!=nil{
		tx.Rollback()
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":validationerr.Error()})
		log.Panicln(validationerr)
	}
	
	inserUserSmt := "INSERT INTO users (first_name, last_name, email) values($1, $2, $3) RETURNING id"
	insertPhoneNumber := "INSERT INTO contacts (number, user_id) values ($1, $2)"
	id := 0
	dberr := tx.QueryRow(inserUserSmt,user.FirstName,user.LastName,user.Email).Scan(&id)
	if dberr != nil{
		tx.Rollback()
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":dberr.Error()})
		log.Panicln(dberr)
		
	}
	// contacts := []Contacts{}
	// valueStrings := []string{}
	// values := []interface{}{}
	// c := Contacts{}
	// 	c.Number = contact
	// 	c.UserId = id
	// 	contacts = append(contacts, c)
	
	for _, contact:= range user.Contacts{
		_ , err := tx.Exec(insertPhoneNumber, contact, id)
		if err!=nil{
			tx.Rollback()
			ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			log.Panicln(err)
			
		}
	}
	tx.Commit()
	ctx.IndentedJSON(http.StatusCreated, "success")
}

func SetupDB() *sql.DB {
	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_name := os.Getenv("DB_NAME")
	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD") 
	
	db_ssl_mode := os.Getenv("DB_SSL_MODE")
	dataSourceURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", db_user, db_password, db_host, db_port, db_name, db_ssl_mode)
	fmt.Println(dataSourceURL)

	db, err := sql.Open("postgres", dataSourceURL)
	checkErr(err, "sql.Open failed")
	dberr := db.Ping()
	checkErr(dberr,"failed to connect to db")
	return db
  }

  func checkErr(err error, msg string) {
	if err != nil {
		log.Panicln(msg, err)
	}
}


func exitErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}


func main(){
	
	err := godotenv.Load()
	exitErr(err,"Error loading .env file")
	router := gin.Default()
	handlers := gin.New()
	handlers.Use(gin.Recovery())
	router.LoadHTMLGlob("ui/*")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})
	router.GET("/contacts", getUsers)
	router.POST("/contacts", addUsers)
	server_host := os.Getenv("SERVER_HOST")
	server_port := os.Getenv("SERVER_PORT")
	if server_host == "" || server_port == "" {
		log.Fatal("missing server configs")
	}
	router.Run(fmt.Sprintf("%s:%s",server_host,server_port))
	
}