package main

// Using Go Lang gin as a HTTP web framework
// godotenv to load db and server environment configuration
// libphonenumber a google library for parcing mobile numbers
// postgres db driver

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

// User contacts to be stored in relation database with web interface to insert and retreve data
// Ensure to add create tables added in the data.sql to the database schema
// database is normalized to sotre user contact information in to two tables users & contacts 
// Assumptions:
// Every user will have unique email id, could have 1 to n number of mobile numbers


// GET usercontacts model structure
type getUserRequest struct {
	Name string `json:"full_name"`
	Email string `json:"email"`
	Contacts []string `json:"phone_numbers"`
}
// POST usercontacts model structure
type postUserRequest struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Email string `json:"email"`
	Contacts []string `json:"phone_numbers"`
}


// A get request to fetch the list of users with their contact related information

func getUsers(ctx *gin.Context) {
	// initiate db connection
	db := SetupDB()
	// Todo: use postgres ORM to perform database operations
	// I choose to use raw sql statements for this demo test purposes
	// Following SQL query joins the users & contacts tables on the reference relation user id,
	// group by email(unique filed) and aggregating the contact numbers into list 
	rows, err := db.Query("SELECT  CONCAT(users.first_name , ' ' ,users.last_name) as Name, users.email as Email, ARRAY_AGG(contacts.number) as Contacts FROM users JOIN contacts on users.id=contacts.user_id  GROUP BY Name, Email ")
	checkErr(err, "error while querying users")
	defer rows.Close()
	users := []getUserRequest{}

	for rows.Next(){
		checkErr(rows.Err(),"Error while getting db rows")
		u := getUserRequest{}
		// fetching the resultset from database query
		err := rows.Scan(&u.Name, &u.Email, pq.Array(&u.Contacts))
		checkErr(err,"Error while scanning user data")
		users = append(users, u)
	}
	// returning 200 status with users contact information as a list
	ctx.IndentedJSON(http.StatusOK, users)

}

// A Post request context function that will create new user contact entry
// Incase of error falure gracesfully logs the error to console and continues to serve

func addUsers(ctx *gin.Context){
	
	// initiate db connection
	db := SetupDB()
	
	// Transaction statement intitiation
	// As we  are inserting new record in to db table user has forign key reference for storing contact records
	// As mentioned in the assumption one user may have multiple contact number, contacts is a normalized talble to store user mobile numbers
	
	// Todo: Although following db transaction functinality works as expected but there is a room for improvement
	// when hadling the reference data 
	tx, txerr := db.Begin()
	if txerr!=nil{
		log.Panicln(txerr)
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":txerr.Error()})
	}

	user := postUserRequest{}
	// binding the request data with the interface
	ctx.BindJSON(&user)

	// Todo: Improve following error handling
	// Though I have added the minimal validation required at client level
	// I want to ensure API request is properly validated and handled the exception
	// However endup writing validateRequest function which serves the purpost but could have improved

	_, validationerr:=validateRequest(user)
	if validationerr!=nil{
		tx.Rollback()
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":validationerr.Error()})
		log.Panicln(validationerr)
	}
	// insert statement for creating new user
	inserUserSmt := "INSERT INTO users (first_name, last_name, email) values($1, $2, $3) RETURNING id"
	// insert statement for adding new user contact number
	insertPhoneNumber := "INSERT INTO contacts (number, user_id) values ($1, $2)"
	id := 0
	dberr := tx.QueryRow(inserUserSmt,user.FirstName,user.LastName,user.Email).Scan(&id)
	// gracefully handling the exception
	if dberr != nil{
		tx.Rollback()
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":dberr.Error()})
		log.Panicln(dberr)
		
	}
	// Todo: did not manage to implement bulk insert but works for the task
	for _, contact:= range user.Contacts{
		_ , err := tx.Exec(insertPhoneNumber, contact, id)
		if err!=nil{
			tx.Rollback()
			ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			log.Panicln(err)
			
		}
	}
	// finally! commiting the transaction changs if no errors
	tx.Commit()
	ctx.IndentedJSON(http.StatusCreated, "success")
}



func validateRequest( user postUserRequest) (bool,error){
	// Email validation
	if user.Email == "" || ! strings.Contains(user.Email,"@"){
		return false, errors.New("Incorrect or missing email id")
	}
	// First Name validation
	if user.FirstName == "" {
		return false, errors.New("Missing first name")
	}
	if user.LastName == "" {
		return false, errors.New("Missing Last name")
	}
	// validation user contacts
	// Todo: improve error handling
	if len(user.Contacts) == 0 {
		return false, errors.New("Empty phone_numbers list")
	}else if len(user.Contacts) > 0  {
		// Validation contact numbers with google port in libphonenumber golang library
		// prodived dynamic parsing with respect to regioun provided
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



// db serup config initiation file
// returns open connection db object that can be used to interact with the database
// exits code with error message incase of failure
func SetupDB() *sql.DB {
	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_name := os.Getenv("DB_NAME")
	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWORD") 
	
	db_ssl_mode := os.Getenv("DB_SSL_MODE")
	dataSourceURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", db_user, db_password, db_host, db_port, db_name, db_ssl_mode)

	db, err := sql.Open("postgres", dataSourceURL)
	exitErr(err, "sql.Open failed")

	// validating database config provided
	// incase of failure exists code with error message
	dberr := db.Ping()
	exitErr(dberr,"failed to connect to db")
	return db
  }

// error handling function to log the error with addition of custome message
//  to keep service recovered from error and capture the error related info
func checkErr(err error, msg string) {
	if err != nil {
		log.Panicln(msg, err)
	}
}

// error handling function to exit the running code 
func exitErr(err error, msg string) {
	if err != nil {
		// logs the error with custome message before exit
		log.Fatalln(msg, err)
	}
}


func main(){
	
	// loads the .env file from project root folder
	// incase of failure to load .env througs fatal error

	err := godotenv.Load()
	exitErr(err,"Error loading .env file")

	//initiates and creates a gin router with default middleware
	router := gin.Default()
	//initiating gin recovery for keeping server active incase of exceptions
	handlers := gin.New()
	handlers.Use(gin.Recovery())
	
	// LoadHTMLGlob to load the index file
	// and associates the result with HTML renderer.
	router.LoadHTMLGlob("ui/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Hello World",
		})
	})
	// GET API call to fetch user contact information
	router.GET("/contacts", getUsers)
	// POST API call to add new user contact information
	router.POST("/contacts", addUsers)

	//loading server environment files
	//incase of incorrect configueration throughs fatal error with exit code
	server_host := os.Getenv("SERVER_HOST")
	server_port := os.Getenv("SERVER_PORT")
	if server_host == "" || server_port == "" {
		log.Fatal("missing server configs")
	}
	// listen and serve on given host & port
	router.Run(fmt.Sprintf("%s:%s",server_host,server_port))
	
}