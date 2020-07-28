
package webClient

import (
	// "strings"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"github.com/pkg/errors"
)


// appTemplate is a user login-aware wrapper for a html/template.
type appTemplate struct {
	t *template.Template
}
func AppURL(smid string) string{
	return "/feeds/ws?appid="+smid
}

// parseTemplate applies a given file to the body of the base template.
func NewAppTemplate(filename string) *appTemplate {
	fm := template.FuncMap{
		"appUrl": AppURL,
	}
	tmpl := template.Must(template.ParseFiles("webClient/templates/base.html"))
	// var tmpl *template.Template
	// if filename == "index.html" {
	// 	//Read in base.html as a string to remove sidebar template {{}} from the index page
	// 	b, err := ioutil.ReadFile("webClient/templates/base.html")
	// 	if err != nil {
	// 		panic(fmt.Errorf("could not read template: %v", err))
	// 	}
	// 	ba := strings.Replace(string(b), "{{template \"sidebar\" .User}}", "", 1)
	// 	ba = strings.Replace(string(ba), "w3-hide-small\" id=\"forIndex\"", "\" id=\"forIndexA\"", 3)
	// 	//So now the stringified base.html is tranforming back to be a template without the sidebar
	// 	tmpl = template.Must(template.New("base").Parse(ba))
	// 	tmpl = template.Must(tmpl.ParseFiles("webClient/templates/chat.html"))
	// }else {
	// 	tmpl = template.Must(template.ParseFiles("webClient/templates/base.html", "webClient/templates/chat.html", "webClient/templates/sidebar.html"))
	// }
	// Put the named file into a template called "body"
	path := filepath.Join("webClient/templates", filename)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("could not read template: %v", err))
	}
	template.Must(tmpl.New("body").Funcs(fm).Parse(string(b)))
	return &appTemplate{tmpl.Lookup("base.html")}
}

// Execute writes the template using the provided data, adding login and user
// information to the base template.
func (tmpl *appTemplate) Execute(w http.ResponseWriter, r *http.Request, data interface{}, usr interface{}, msg interface{}) error {
	d := struct {
		Data        interface{}
		LoginURL    string
		LogoutURL   string
		//AddFooter   bool
		SignupURL string
		User      interface{}
		//Rate      string
		Msg       interface{}
	}{
		Data:        data,
		LoginURL:    "/login?redirect=" + r.URL.RequestURI(),
		LogoutURL:   "/logout?redirect=" + r.URL.RequestURI(),
		SignupURL:   "/signup?redirect=" + r.URL.RequestURI(),
		//AddFooter:   noFooter,
		User: usr,
		//Rate: <-RateChan,
		Msg: msg,
	}
	if err := tmpl.t.Execute(w, d); err != nil {
		return errors.Wrapf(err, "could not write template: %+v", err)
	}
	return nil
}
