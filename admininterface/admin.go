package admininterface

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

var Signuperror models.Invalidsignup

var userNameToUpdate string

var new models.Signupusers

var c http.Cookie

var err error

var Userslist []models.Signupusers

func Adminhandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	_, err := r.Cookie("jwt_admin_token")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	database.Db.Find(&Userslist)
	data := models.Searchdata{
		Users: Userslist,
	}

	if len(Userslist) == 0 {
		data.SearchError = "No users found"
	}

	tmp, err := template.ParseFiles("templates/admin/admin.html")
	if err != nil {
		log.Fatalf("error %v", err)
	}
	tmp.ExecuteTemplate(w, "admin.html", data)

}

func Adminlogout(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {

		c = http.Cookie{Name: "jwt_admin_token", Value: "", Expires: time.Now().AddDate(0, 0, -1), MaxAge: -1}
		http.SetCookie(w, &c)

		http.Redirect(w, r, "/", http.StatusSeeOther)

	} else {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

}

func MiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("jwt_admin_token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Missing authorization cookie")
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error retrieving cookie: %v", err)
			return
		}

		tokenString := cookie.Value
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Missing authorization token")
			return
		}

		token, err := jwtoken.ParseToken(tokenString)
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Invalid authorization token: %v", err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AdminAddUser(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/adminadduser" {
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

		if new.Fullname == "" || new.Email == "" || new.Username == "" || new.Password == "" {

			Signuperror.InvalidFullname = "Invalid data to sign up"
			Signuperror.InvalidEmail = "Invalid data to sign up"
			Signuperror.InvalidUsername = "Invalid data to sign up"
			Signuperror.InvalidPassword = "Invalid data to sign up"

			tmp, err := template.ParseFiles("templates/admin/adminadduser.html")
			if err != nil {
				log.Fatalf("error %v", err)
			}

			tmp.ExecuteTemplate(w, "adminadduser.html", Signuperror)
			return
		}
		if new.Password != r.FormValue("confirmpassword") {
			Signuperror.InvalidFullname = ""
			Signuperror.InvalidEmail = ""
			Signuperror.InvalidUsername = ""
			Signuperror.InvalidPassword = "two password must match"
			tmp, err := template.ParseFiles("templates/admin/adminadduser.html")
			if err != nil {
				log.Fatalf("error %v", err)
			}

			tmp.ExecuteTemplate(w, "adminadduser.html", Signuperror)
			return
		}
		var count int64
		database.Db.Model(&models.Signupusers{}).Where("username = ?", new.Username).Count(&count)
		if count != 0 {
			Signuperror.InvalidFullname = ""
			Signuperror.InvalidEmail = ""
			Signuperror.InvalidUsername = "Already registered Username"
			Signuperror.InvalidPassword = ""

			tmp, err := template.ParseFiles("templates/admin/adminadduser.html")
			if err != nil {
				log.Fatalf("error %v", err)
			}
			tmp.ExecuteTemplate(w, "adminadduser.html", Signuperror)
			return
		}

		query := "INSERT INTO signupusers (fullname,email,username,password) VALUES ($1,$2,$3,$4)"
		err = database.Db.Exec(query, new.Fullname, new.Email, new.Username, new.Password).Error
		if err != nil {
			log.Fatalf("error %v", err.Error())
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)

	}

	tmp, err := template.ParseFiles("templates/admin/adminadduser.html")
	if err != nil {
		log.Fatalf("error %v", err)
	}
	tmp.ExecuteTemplate(w, "adminadduser.html", nil)

}

func Adminuserdelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == "POST" {

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
			return
		}

		database.Db.Where("username", r.FormValue("usingNameToDelete")).Delete(&models.Signupusers{})
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Error(w, "404 not found.", http.StatusNotFound)
	}
}

func AdminUserUpdate(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}
	numValues := len(r.Form)
	fmt.Println("form value number ", numValues)

	if r.Method == "POST" && numValues == 1 {

		var PassData models.UserUpdate
		var user models.Signupusers

		userNameToUpdate = r.FormValue("usingNameToUpdate")

		Signuperror.InvalidFullname = ""
		Signuperror.InvalidEmail = ""
		Signuperror.InvalidPassword = ""

		if err := database.Db.Raw("SELECT * FROM signupusers WHERE username = ?", userNameToUpdate).Scan(&user).Error; err != nil {
			fmt.Println("Error:", err)
			return
		}

		PassData = models.UserUpdate{
			Error:    Signuperror,
			Userdata: user,
		}

		tmp, err := template.ParseFiles("templates/admin/adminuserupdate.html")
		if err != nil {
			log.Fatalf("error %v", err)
		}
		tmp.ExecuteTemplate(w, "adminuserupdate.html", PassData)
		return
	}

	if r.Method == "POST" && numValues > 1 {

		var PassData models.UserUpdate
		var user models.Signupusers

		new.Fullname = r.FormValue("fullname")
		new.Email = r.FormValue("email")
		new.Password = r.FormValue("password")

		fmt.Println("hello")

		if new.Fullname == "" || new.Email == "" || new.Password == "" {

			Signuperror.InvalidFullname = "Invalid data in sign up"
			Signuperror.InvalidEmail = "Invalid data in sign up"
			Signuperror.InvalidPassword = "Invalid data in sign up"

			if err := database.Db.Raw("SELECT * FROM signupusers WHERE username = ?", userNameToUpdate).Scan(&user).Error; err != nil {
				fmt.Println("Error:", err)
				return
			}

			PassData = models.UserUpdate{
				Error:    Signuperror,
				Userdata: user,
			}

			tmp, err := template.ParseFiles("templates/admin/adminuserupdate.html")
			if err != nil {
				log.Fatalf("error %v", err)
			}
			tmp.ExecuteTemplate(w, "adminuserupdate.html", PassData)

		}

		database.Db.Model(&models.Signupusers{}).Where("username = ?", userNameToUpdate).Updates(map[string]interface{}{
			"fullname": new.Fullname,
			"email":    new.Email,
			"password": new.Password,
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}

}

func AdminSearchUser(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		fmt.Println("error here", err)
		http.Error(w, "Failed to parse form data", http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {

		var usersearch []models.Signupusers

		usernametosearch := r.FormValue("usernametosearch")

		query := database.Db.Where("username LIKE ?", "%"+usernametosearch+"%")

		query.Find(&usersearch)
		searchdata := models.Searchdata{
			Users: usersearch,
		}
		if len(usersearch) == 0 {
			searchdata.SearchError = "No users found"
		}

		fmt.Println(searchdata)

		tmp, err := template.ParseFiles("templates/admin/admin.html")
		if err != nil {
			log.Fatalf("error %v", err)
		}
		tmp.ExecuteTemplate(w, "admin.html", searchdata)
	}
}
