package userinterface

import (
	"admincontrol/database"
	"admincontrol/jwtoken"
	"admincontrol/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

var h models.Home

var c http.Cookie

var new models.Signupusers

var Signuperror models.Invalidsignup
var loginerror models.Invalidlogin

var err error

var loginuser string

func RootHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method == "POST" {
		w.Header().Set("Cache-Control", "no-store")

		var validuser string
		var validpassword string

		if err := r.ParseForm(); err != nil {
			fmt.Println("error here", err)
			http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
			return
		}
		loginuser = r.FormValue("username")
		loginpassword := r.FormValue("password")

		if loginuser == "admin" && loginpassword == "sreejith" {
			loginerror.Username = ""
			loginerror.Password = ""
			tokenString, err := jwtoken.GenerateJWT(loginuser, "admin")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error creating token: %v", err)
				return
			}
			cookie := http.Cookie{
				Name:     "jwt_admin_token",
				Value:    tokenString,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "/admin", http.StatusSeeOther)

			return
		}

		var count int64
		database.Db.Model(&models.Signupusers{}).Where("username = ?", loginuser).Count(&count)
		if count != 0 {
			fmt.Println("username exists")
			database.Db.Model(&models.Signupusers{}).Where("username = ?", loginuser).Pluck("username", &validuser)
			database.Db.Model(&models.Signupusers{}).Where("username = ?", loginuser).Pluck("password", &validpassword)

		}

		if loginuser == validuser && loginpassword == validpassword && loginuser != "" && loginpassword != "" {
			loginerror.Username = ""
			loginerror.Password = ""

			tokenString, err := jwtoken.GenerateJWT(loginuser, "user")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error creating token: %v", err)
				return
			}
			cookie := http.Cookie{
				Name:     "jwt_token",
				Value:    tokenString,
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			}
			http.SetCookie(w, &cookie)

			http.Redirect(w, r, "/home", http.StatusSeeOther)

		} else if loginuser != validuser && loginpassword == validpassword {
			loginerror.Username = "Invalid username"
			loginerror.Password = ""

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else if loginuser == validuser && loginpassword != validpassword {
			loginerror.Username = ""
			loginerror.Password = "Invalid password"

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			loginerror.Username = "Invalid username"
			loginerror.Password = "Invalid password"

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}

	}

	_, err := r.Cookie("jwt_token")
	if err == nil {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
	s, err := r.Cookie("jwt_admin_token")
	if err == nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
	fmt.Println(s)

	tmp, err := template.ParseFiles("templates/user/login.html")
	if err != nil {
		log.Fatalf("error %v", err)
	}
	tmp.ExecuteTemplate(w, "login.html", loginerror)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/signup" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Method == "POST" {

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		new.Fullname = r.FormValue("fullname")
		new.Email = r.FormValue("email")
		new.Username = r.FormValue("username")
		new.Password = r.FormValue("password")

		fmt.Println(new.Fullname)

		if new.Fullname == "" || new.Email == "" || new.Username == "" || new.Password == "" {

			Signuperror.InvalidFullname = "Invalid data to sign up"
			Signuperror.InvalidEmail = "Invalid data to sign up"
			Signuperror.InvalidUsername = "Invalid data to sign up"
			Signuperror.InvalidPassword = "Invalid data to sign up"

			tmp, err := template.ParseFiles("templates/user/signup.html")
			if err != nil {
				log.Fatalf("error %v", err)
			}

			tmp.ExecuteTemplate(w, "signup.html", Signuperror)
			return
		}
		if new.Password != r.FormValue("confirmpassword") {
			Signuperror.InvalidFullname = ""
			Signuperror.InvalidEmail = ""
			Signuperror.InvalidUsername = ""
			Signuperror.InvalidPassword = "two password must match"
			tmp, err := template.ParseFiles("templates/user/signup.html")
			if err != nil {
				log.Fatalf("error %v", err)
			}

			tmp.ExecuteTemplate(w, "signup.html", Signuperror)
			return
		}
		var count int64
		database.Db.Model(&models.Signupusers{}).Where("username = ?", new.Username).Count(&count)
		if count != 0 {
			Signuperror.InvalidFullname = ""
			Signuperror.InvalidEmail = ""
			Signuperror.InvalidUsername = "Already registered Username"
			Signuperror.InvalidPassword = ""

			tmp, err := template.ParseFiles("templates/user/signup.html")
			if err != nil {
				log.Fatalf("error %v", err)
			}
			tmp.ExecuteTemplate(w, "signup.html", Signuperror)
			return
		}

		query := "INSERT INTO signupusers (fullname,email,username,password) VALUES ($1,$2,$3,$4)"
		err = database.Db.Exec(query, new.Fullname, new.Email, new.Username, new.Password).Error
		if err != nil {
			log.Fatalf("error %v", err.Error())
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	}

	tmp, err := template.ParseFiles("templates/user/signup.html")
	if err != nil {
		log.Fatalf("error %v", err)
	}
	tmp.ExecuteTemplate(w, "signup.html", nil)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	_, err := r.Cookie("jwt_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	h.Username = loginuser

	tmp, err := template.ParseFiles("templates/user/home.html")
	if err != nil {
		log.Fatalf("error %v", err)
	}
	tmp.ExecuteTemplate(w, "home.html", h)

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		c = http.Cookie{Name: "jwt_token", Value: "", Expires: time.Now().AddDate(0, 0, -1), MaxAge: -1}

		http.SetCookie(w, &c)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "404 not found.", http.StatusNotFound)
	}
}
