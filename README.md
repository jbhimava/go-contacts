# GO Lang Contacts demo application
Demo Web Application for Adding user contact information

A simple web application with Go using

1. GIN Framework
2. Postgres
3. HTML/JavaScript

![Demo App](demo_app.png)

## API List

### GET : Get all contacts
```
GET /contacts HTTP/1.1
Host: localhost:8080
```
### POST : Create a contact
```
POST /contacts HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{ "first_name":"first_name","last_name":"last_name","email":"email@email.com","phone_numbers":["04********", "+614********"]}
```
## Project setup
* Install latest version  [Go Lang](https://golang.org/dl/)
* Install latest version of [Postgres](https://www.postgresql.org/download/) database 
* Optional install [Visual Studio Code](https://code.visualstudio.com/docs/languages/go) for GO Lang

## Quick Start

1. Clone GO-Contanct repo to your local and add environment variables required , rename `.sample_env` to `.env`  
2. Add users & contacts tables to your DB schema by running the create table script from [data.sql](data.sql)
3. After updating the `.env` file run the following commands from root folder

```
go mod tidy
go run .
```
> All set to add and retrieve user contact information using go web server

## Useful resources
- [GO Lang](https://golang.org/doc/tutorial/getting-started)
- [GO Lang with GIN for REST API](https://golang.org/doc/tutorial/web-service-gin)
- [Postgres](https://pkg.go.dev/github.com/lib/pq)
- [Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API)

## Todo
* More elegance config
* Add Test Coverage
* Improve error handling
* Project/code structure optimize
* Add Build scripts


