package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Result struct {
	Code    int         `json="code"`
	Message string      `json="message"`
	Data    interface{} `json="data"`
}
type Info struct {
	Github string
	Office string
	Phone  string
}
type Personal struct {
	Name      string `json="name"`
	Email     string `json="email"`
	Address   string `json="address"`
	ShortName string `json="shortname"`
	Age       int    `json="age"`
	Info      Info
	Addr      map[string]Address
	Hobbies   []Hobbie
}

func (u Personal) HasPermission(feature string) bool {
	if feature == "feature-a" {
		return true
	} else {
		return false
	}
}

type Address struct {
	Name        string
	Description string
}
type Hobbie struct {
	Name        string
	Description string
}

func main() {
	http.Handle("/static/",
		http.StripPrefix("/static/", // mengahapus prefix dari endpoint yang mengarah ke static/assets
			http.FileServer(http.Dir("assets"))))
	//router
	http.HandleFunc("/contact", contact)
	http.HandleFunc("/", index)
	http.HandleFunc("/about", about)
	http.HandleFunc("/contributer", contributer)
	http.HandleFunc("/send-message", sendMessage)
	http.HandleFunc("/upload-portfolio", uploadPortfolio)
	http.HandleFunc("/process-upload", processUpload)
	fmt.Println("server started at http://localhost:8080")
	_ = http.ListenAndServe(":8080", nil)
}

func processUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST"{
		http.Error(w, "yang diperbolehkan hanya post", http.StatusInternalServerError)
		return
	}

	// ambil multiform/formdata
	err := r.ParseMultipartForm(1024)// max memory
	if isError(err){
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	alias:= r.FormValue("alias")
	uploadedFile, handler, err :=r.FormFile("file")
	if isError(err){
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer uploadedFile.Close()

	// get absolute path dimana app di run
	dir, err := os.Getwd()
	if isError(err){
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filename := handler.Filename
	if alias != ""{
		filename = fmt.Sprintf("%s%s", alias, filepath.Ext(filename))
	}
	fileLocation := filepath.Join(dir, "files", filename )
	targetFile, err:= os.OpenFile(fileLocation, os.O_WRONLY | os.O_CREATE, 0666)
	if isError(err){
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer targetFile.Close()

	if _, err = io.Copy(targetFile, uploadedFile); err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte("done"))
}

func uploadPortfolio(w http.ResponseWriter, r *http.Request) {
	var tmpl = template.Must(template.ParseFiles("views/upload-portfolio.html", "views/_header.html", "views/_footer.html"))
	var err = tmpl.ExecuteTemplate(w, "upload-portfolio", nil)
	if  isError(err){
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func sendMessage(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "GET" {
		http.Error(writer, "Hanya diperbolehkan GET", http.StatusBadRequest)
		return
	}
	var subject = request.FormValue("subject")
	var email = request.Form.Get("email")
	var jsonRes = request.Form.Get("json")
	var deskripsi = request.FormValue("message")
	var result = map[string]string{
		"subject":   subject,
		"email":     email,
		"message": deskripsi,
	}
	if jsonRes != "" {

		jsonByte, err := json.Marshal(&result)
		_, err = writer.Write(jsonByte)
		if isError(err) {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
	tmpl := template.Must(template.ParseFiles("views/result-message.html", "views/_header.html", "views/_footer.html"))
	_ = tmpl.ExecuteTemplate(writer, "result-message", result)

}

// function map
var funcMap = template.FuncMap{
	"add": func(val, addNumber int) int {
		return val + addNumber
	},
}

func contributer(w http.ResponseWriter, request *http.Request) {
	var person = Personal{
		Name:  "Zaza zayinul hikayat",
		Email: "dzas42@gmail.com",
	}
	person.Hobbies = []Hobbie{{"Codding", "fun and cool "}, {"Watching", "Watching in bioskop"}}
	var tmpl = template.Must(template.New("contributer").
		Funcs(funcMap).ParseFiles("views/contributer.html"))
	if err := tmpl.Execute(w, person); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func about(writer http.ResponseWriter, request *http.Request) {
	var err error
	var tmpl = template.Must(template.ParseFiles(
		"views/about.html", "views/_header.html", "views/_footer.html"))

	err = tmpl.ExecuteTemplate(writer, "about", nil)

	if isError(err) {
		panic(err.Error())
	}
}

func contact(writer http.ResponseWriter, request *http.Request) {
	// template akan merender semua file yang ada di dalam folder views/*
	var tmpl, err = template.ParseGlob("views/*")
	tmpl.Funcs(funcMap)
	var me Personal = Personal{Name: "Zaza zayinul hikayat", Email: "dzas42@gmail.com", Info: Info{Github: "http://github/dzas42"},
		Addr: map[string]Address{"home": {"Home", "Jakarta Timur"}, "office": {"office", "Jakarta Pusat"}},
	}
	err = tmpl.ExecuteTemplate(writer, "contact", me)

	if isError(err) {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		//var result Result
		//result = Result{Code: http.StatusBadGateway, Message: err.Error(),
		//}
		//jsonval, _ := json.Marshal(result)
		//setResponseJson(writer, jsonval)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	var msg = Result{http.StatusOK, "welcome web api", []string{}}
	json, _ := json.Marshal(msg)
	setResponseJson(w, json)
}

func setResponseJson(w http.ResponseWriter, msg []byte) {
	w.Header().Set("content-type", "application/json ")
	_, _ = w.Write(msg)
}
func isError(err error) bool {
	if err != nil {
		fmt.Println("Terdapat error ", err.Error())
		return true
	}
	return false
}
