package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
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

func getBookFilePath(bookFileName string) string {
	return "./books/" + bookFileName + ".json"
}

func loadBook(fileName string) Book {
	filePath := getBookFilePath(fileName)
	body, _ := ioutil.ReadFile(filePath)

	var book Book
	json.Unmarshal([]byte(body), &book)
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

// TODO clean this up
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

	Header := r.FormValue("Header")
	Caption := r.FormValue("Caption")
	ImageUrl := r.FormValue("ImageUrl")
	Body := r.FormValue("body")

	book := loadBook(bookFileName)

	var newMilestones []Milestone
	if milestoneId > 0 {
		for _, milestone := range book.Milestones {
			if milestone.Id == milestoneId {
				milestone.Header = Header
				milestone.Caption = Caption
				milestone.ImageUrl = ImageUrl
				milestone.ImageUrl = ImageUrl
				milestone.Body = Body
			}
			newMilestones = append(newMilestones, milestone)
		}

	} else {

		newMilestone := Milestone{
			Id:       len(book.Milestones) + 1,
			Order:    len(book.Milestones) + 1,
			Header:   Header,
			Caption:  Caption,
			ImageUrl: ImageUrl,
			Body:     Body,
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

	for _, m := range book.Milestones {
		fmt.Printf("%+v\n", m)
	}

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

func main() {
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/book/", bookHandler)
	http.HandleFunc("/add_or_edit_milestone/", addOrEditMilestoneHandler)
	http.HandleFunc("/save_milestone/", saveMilestoneHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
