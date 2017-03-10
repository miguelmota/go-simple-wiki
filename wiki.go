package main

import (
  "os"
  "log"
  "io/ioutil"
  "net/http"
  "html/template"
  "regexp"
  "errors"
)

type Page struct {
  Title string;
  Body []byte
}

// cache templates
var templates = template.Must(template.ParseFiles("./tmpl/edit.html", "./tmpl/view.html"))

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
  m := validPath.FindStringSubmatch(r.URL.Path)
  if m == nil {
    http.NotFound(w, r)
    return "", errors.New("Invalid page title")
  }
  return m[2], nil
}

func getFilePath(title string) (string) {
  return "./data/" + title + ".txt"
}

func (p *Page) save() error {
  return ioutil.WriteFile(getFilePath(p.Title), p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
  body, err := ioutil.ReadFile(getFilePath(title))
  if err != nil {
    return nil, err
  }
  return &Page{Title: title, Body: body}, nil
}

func makeHandler(fn func(w http.ResponseWriter, r *http.Request, title string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/" + title, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

func main() {
  log.SetOutput(os.Stdout)

  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))

  host := ":8080"
  log.Println("Listening on " + host)
  http.ListenAndServe(host, nil)
}