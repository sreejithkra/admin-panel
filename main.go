package main

import (
	"admincontrol/admininterface"
	"admincontrol/database"
	"admincontrol/userinterface"
	"fmt"
	"log"
	"net/http"
)

func main() {

	fmt.Println("hello")

	database.Gormconnect()

	http.HandleFunc("/", userinterface.RootHandler)
	http.HandleFunc("/signup", userinterface.SignupHandler)
	http.HandleFunc("/home", userinterface.HomeHandler)
	http.HandleFunc("/logout", userinterface.LogoutHandler)
	http.HandleFunc("/admin", admininterface.Adminhandler)
	http.HandleFunc("/adminlogout", admininterface.Adminlogout)
	http.Handle("/adminadduser", admininterface.MiddleWare(http.HandlerFunc(admininterface.AdminAddUser)))
	http.Handle("/adminUserDelete", admininterface.MiddleWare(http.HandlerFunc(admininterface.Adminuserdelete)))
	http.Handle("/adminUserUpdate", admininterface.MiddleWare(http.HandlerFunc(admininterface.AdminUserUpdate)))
	http.Handle("/adminSearchUser", admininterface.MiddleWare(http.HandlerFunc(admininterface.AdminSearchUser)))

	fmt.Printf("Starting server at port 8086\n")
	if err := http.ListenAndServe(":8086", nil); err != nil {
		log.Fatal(err)
	}
}
