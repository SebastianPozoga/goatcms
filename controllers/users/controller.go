package users

import (
	"errors"
	"log"
	"net/http"

	"github.com/goatcms/goat-core/dependency"
	"github.com/goatcms/goatcms/models"
	"github.com/goatcms/goatcms/models/user"
	"github.com/goatcms/goatcms/services"
)

// UserController is article controller endpoint
type UserController struct {
	tmpl    services.Template
	userDAO models.UserDAO
	crypt   services.Crypt
}

// NewUserController create instance of a articles controller
func NewUserController(dp dependency.Provider) (*UserController, error) {
	ctrl := &UserController{}
	// load template service from dependency provider
	tmplIns, err := dp.Get(services.TemplateID)
	if err != nil {
		return nil, err
	}
	ctrl.tmpl = tmplIns.(services.Template)
	// load userDAO service from dependency provider
	daoIns, err := dp.Get(models.UserDAOID)
	if err != nil {
		return nil, err
	}
	ctrl.userDAO = daoIns.(models.UserDAO)
	// load crypting service from dependency provider
	cryptIns, err := dp.Get(services.CryptID)
	if err != nil {
		return nil, err
	}
	ctrl.crypt = cryptIns.(services.Crypt)
	return ctrl, nil
}

// ValidationError type for form validation errors
type ValidationError error

var (
	errNoEmail      = ValidationError(errors.New("You must supply an email"))
	errNoPassword   = ValidationError(errors.New("You must supply a password"))
	errPassTooShort = ValidationError(errors.New("Password should be at least 8 characters"))
)

// IsValidationError check if given error is of type ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(ValidationError)
	return ok
}

const (
	passwordLength = 6
)

func (c *UserController) newUser(email, pass string) (usermodel.UserDTO, error) {
	user := usermodel.UserDTO{
		Email:    email,
		PassHash: pass,
	}
	if email == "" {
		return user, errNoEmail
	}
	if pass == "" {
		return user, errNoPassword
	}
	if len(pass) < passwordLength {
		return user, errPassTooShort
	}
	// perform password hashing before returning UserDTO object
	passHashed, err := c.crypt.Hash(pass)
	if err != nil {
		return user, err
	}
	user.PassHash = passHashed

	return user, err
}

// TemplateSignUp is handler to serve template where one can register new user
func (c *UserController) TemplateSignUp(w http.ResponseWriter, r *http.Request) {
	log.Println("responding to", r.Method, r.URL)
	err := c.tmpl.ExecuteTemplate(w, "users/register", nil)
	if err != nil {
		log.Fatal("error rendering a template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// TryToSignUp is handler to save user from form obtained data
func (c *UserController) TryToSignUp(w http.ResponseWriter, r *http.Request) {
	log.Println("responding to", r.Method, r.URL)
	// TODO: http://www.gorillatoolkit.org/pkg/schema
	// like: err := decoder.Decode(person, r.PostForm)
	// By Sebastian
	err := r.ParseForm()
	if err != nil {
		log.Fatal("error parsing a form: ", err)
		return
	}
	// obtain data from form
	email := r.PostFormValue("email")
	passPlaintext := r.PostFormValue("password")
	// validate inputs
	user, err := c.newUser(email, passPlaintext)
	if err != nil {
		if IsValidationError(err) {
			c.tmpl.ExecuteTemplate(w, "users/register", map[string]interface{}{
				"Error": err.Error(),
				"Email": user.GetEmail(),
			})
			return
		}
		panic(err)
	}
	// encrypt password with bcrypt
	// passHashed, err := c.crypt.Hash(passPlaintext)
	// if err != nil {
	// 	log.Fatal("error crypting pass: ", err)
	// 	return
	// }
	// user := usermodel.UserDTO{Email: email, PassHash: passHashed}
	// ...and save to database
	var userToAdd []models.UserDTO
	userToAdd = append(userToAdd, models.UserDTO(&user))
	c.userDAO.PersistAll(userToAdd)
	// redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// TemplateLogin is handler to serve template where one can log in
func (c *UserController) TemplateLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("responding to", r.Method, r.URL)
	err := c.tmpl.ExecuteTemplate(w, "users/login", nil)
	if err != nil {
		log.Fatal("error rendering a template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// TryToLogin is handler to try sign in with given data in POST request
func (c *UserController) TryToLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("responding to", r.Method, r.URL)
	err := r.ParseForm()
	if err != nil {
		log.Fatal("error parsing a form: ", err)
	}
	// obtain data from login form...
	email := r.PostFormValue("email")
	passPlaintext := r.PostFormValue("password")
	// ...and check if that user exist and compare pass with hash from DB
	user := c.userDAO.FindByEmail(email)
	if user == nil {
		log.Println("no users found with email:", email)
	} else {
		result, err := c.crypt.Compare(user.GetPassHash(), passPlaintext)
		if err != nil { // here error means: hash and pass are not matching
			log.Println("password wrong - log in failure")
		}
		if result == true { // if error == nil and result == true
			log.Println("password correct - log in success")
		}
	}
	// redirect
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
