package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

type Milestone struct {
	Id       int
	Order    int
	Header   string
	Caption  string
	ImageUrl string
	Body     string
}

type Book struct {
	FileName   string
	Title      string
	ImageUrl   string
	Authors    []string
	Milestones []Milestone
}

type TemplateAddOrEditMilestoneValue struct {
	Book          Book
	PrevMilestone Milestone
	Milestone     Milestone
}

type TemplateAddOrEditBookValue struct {
	FileName    string
	Title       string
	ImageUrl    string
	AuthorsList string
}

func getBookFilePath(bookFileName string) string {
	return "./books/" + bookFileName + ".json"
}

func loadBook(fileName string) Book {
	filePath := getBookFilePath(fileName)
	body, err := ioutil.ReadFile(filePath)

	var book Book
	if err == nil {
		json.Unmarshal([]byte(body), &book)
	}

	return book
}

func (book *Book) save() error {
	filePath := getBookFilePath(book.FileName)
	b, _ := json.Marshal(book)
	return ioutil.WriteFile(filePath, b, 0600)
}

func renderTemplate(w http.ResponseWriter, fileName string, book *Book) {
	t, _ := template.ParseFiles(fileName + ".html")
	t.Execute(w, book)
}

func addOrEditBookHandler(w http.ResponseWriter, r *http.Request) {
	bookFileNameQueryValues := r.URL.Query()["bookFileName"]

	var bookFileName string
	if len(bookFileNameQueryValues) > 0 {
		bookFileName = bookFileNameQueryValues[0]
	}

	book := loadBook(bookFileName)

	item := TemplateAddOrEditBookValue{
		FileName:    book.FileName,
		Title:       book.Title,
		ImageUrl:    book.ImageUrl,
		AuthorsList: strings.Join(book.Authors, ","),
	}

	t, _ := template.ParseFiles("add_or_edit_book.html")
	t.Execute(w, item)
}

func saveBookHandler(w http.ResponseWriter, r *http.Request) {
	bookFileNameQueryValues := r.URL.Query()["bookFileName"]

	var bookFileName string
	if len(bookFileNameQueryValues) > 0 {
		bookFileName = bookFileNameQueryValues[0]
	}

	fileName := r.FormValue("FileName")
	title := r.FormValue("Title")
	imageUrl := r.FormValue("ImageUrl")

	authorsList := r.FormValue("AuthorsList")
	authors := strings.Split(authorsList, ",")

	book := loadBook(bookFileName)

	var milestones []Milestone = nil

	if book.FileName == "" {
		milestones = book.Milestones
	}

	book = Book{
		FileName:   fileName,
		Title:      title,
		ImageUrl:   imageUrl,
		Authors:    authors,
		Milestones: milestones,
	}
	book.save()
	http.Redirect(w, r, "/book/"+book.FileName, http.StatusFound)
}

func addOrEditMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	bookFileName := r.URL.Path[len("/add_or_edit_milestone/"):]

	previouisIds := r.URL.Query()["previouisId"]
	milestoneIds := r.URL.Query()["milestoneId"]

	var prevMilestoneId int
	if len(previouisIds) > 0 {
		prevMilestoneId, _ = strconv.Atoi(previouisIds[0])
	}

	var milestoneId int
	if len(milestoneIds) > 0 {
		milestoneId, _ = strconv.Atoi(milestoneIds[0])
	}

	book := loadBook(bookFileName)

	var milestone Milestone
	var prevMilestone Milestone

	for _, m := range book.Milestones {
		if m.Id == milestoneId {
			milestone = m
		}
		if m.Id == prevMilestoneId {
			prevMilestone = m
		}
	}

	item := TemplateAddOrEditMilestoneValue{
		Book:          book,
		PrevMilestone: prevMilestone,
		Milestone:     milestone,
	}

	t, _ := template.ParseFiles("add_or_edit_milestone.html")
	t.Execute(w, item)
}

func saveMilestoneHandler(w http.ResponseWriter, r *http.Request) {
	bookFileName := r.URL.Path[len("/save_milestone/"):]

	previouisIds := r.URL.Query()["previouisId"]
	milestoneIds := r.URL.Query()["milestoneId"]

	var prevMilestoneId int
	if len(previouisIds) > 0 {
		prevMilestoneId, _ = strconv.Atoi(previouisIds[0])
	}

	var milestoneId int
	if len(milestoneIds) > 0 {
		milestoneId, _ = strconv.Atoi(milestoneIds[0])
	}

	header := r.FormValue("Header")
	caption := r.FormValue("Caption")
	imageUrl := r.FormValue("ImageUrl")
	body := r.FormValue("body")

	book := loadBook(bookFileName)

	var newMilestones []Milestone
	if milestoneId > 0 {
		for _, milestone := range book.Milestones {
			if milestone.Id == milestoneId {
				milestone.Header = header
				milestone.Caption = caption
				milestone.ImageUrl = imageUrl
				milestone.Body = body
			}
			newMilestones = append(newMilestones, milestone)
		}

	} else {

		newMilestone := Milestone{
			Id:       len(book.Milestones) + 1,
			Order:    len(book.Milestones) + 1,
			Header:   header,
			Caption:  caption,
			ImageUrl: imageUrl,
			Body:     body,
		}

		pivotOrder := len(book.Milestones)

		if prevMilestoneId > 0 {
			var previousMilestone Milestone
			for _, milestone := range book.Milestones {
				if milestone.Id == prevMilestoneId {
					previousMilestone = milestone
					break
				}
			}

			if (Milestone{} != previousMilestone) {
				pivotOrder = previousMilestone.Order
			} else {
				fmt.Println("Previous milestone not found")
			}
		}

		for _, milestone := range book.Milestones {
			if milestone.Order > pivotOrder {
				milestone.Order += 1
			}
			newMilestones = append(newMilestones, milestone)
		}
		newMilestone.Order = pivotOrder + 1
		newMilestones = append(newMilestones, newMilestone)
	}

	book.Milestones = newMilestones
	book.save()
	http.Redirect(w, r, "/book/"+bookFileName, http.StatusFound)
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	bookFileName := r.URL.Path[len("/book/"):]
	book := loadBook(bookFileName)
	sort.SliceStable(book.Milestones, func(i, j int) bool {
		return book.Milestones[i].Order < book.Milestones[j].Order
	})
	renderTemplate(w, "book", &book)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./books")
	if err != nil {
		log.Fatal(err)
	}
	var allBooks []Book
	for _, f := range files {
		book := loadBook(strings.Split(f.Name(), ".json")[0])
		allBooks = append(allBooks, book)
	}

	t, _ := template.ParseFiles("index.html")
	t.Execute(w, &allBooks)
}

func main() {
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/index/", indexHandler)

	// Books
	http.HandleFunc("/book/", bookHandler)
	http.HandleFunc("/add_or_edit_book/", addOrEditBookHandler)
	http.HandleFunc("/save_book/", saveBookHandler)

	// Milestones
	http.HandleFunc("/add_or_edit_milestone/", addOrEditMilestoneHandler)
	http.HandleFunc("/save_milestone/", saveMilestoneHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
