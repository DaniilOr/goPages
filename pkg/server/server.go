package server

import (
	"encoding/json"
	"github.com/DaniilOr/goPages/pkg/page"
	"github.com/DaniilOr/gorest/pkg/remux"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)
const (
	CHANGED = 201
	DELETED = 204
)
type Server struct{
	Mu sync.RWMutex
	Mux    * remux.ReMUX
	pages []*page.Page
	maxPageID int64
}
func NewService()*Server{
	mu := remux.CreateNewReMUX()

	return &Server{ Mu: sync.RWMutex{}, Mux: mu, pages: []*page.Page{}, maxPageID: 0}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Mux.ServeHTTP(w, r)
}
func (s*Server) GetAll(w http.ResponseWriter, r*http.Request){
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	pages := make([]page.PageDTO, 0, len(s.pages))
	for _, currPage := range s.pages{
		pg := page.PageDTO{Id: currPage.Id, Name: currPage.Name, Img: currPage.Img, Date: currPage.Date}
		pages = append(pages, pg)
	}
	respBody, err := json.Marshal(pages)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		makeResponse(respBody, w, r)
	}
}
func (s*Server) GetSingle(w http.ResponseWriter, r*http.Request){
	s.Mu.Lock()
	defer s.Mu.Unlock()
	var result page.Page
	found := false
	params, err := remux.PathParams(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	id, err := strconv.ParseInt(params.Named["Id"], 10, 64)
	if err != nil{
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	for _, pg := range s.pages{
		if id == pg.Id {
			result.Id = pg.Id
			result.Name = pg.Name
			result.Img = pg.Img
			result.Date = pg.Date
			result.Text = pg.Text
			found = true
			break
		}
	}
	if !found{
		res := page.Result{Result: "Error", ErrorDescription: "Not found"}
		respBody, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			makeResponse(respBody, w, r)
			return
		}
	}
	respBody, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		makeResponse(respBody, w, r)
	}
}
func (s*Server) Change(w http.ResponseWriter, r*http.Request){
	s.Mu.Lock()
	defer s.Mu.Unlock()
	found := false
	params, err := remux.PathParams(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	id, err := strconv.ParseInt(params.Named["Id"], 10, 64)
	if err != nil{
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	var pageId int64
	for i, pg := range s.pages{
		if id == pg.Id {
			pageId = int64(i)
			found = true
			break
		}
	}
	if !found{
		res := page.Result{Result: "Error", ErrorDescription: "Not found"}
		respBody, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			makeResponse(respBody, w, r)
			return
		}
	}
	err = r.ParseForm()
	if err != nil{
		log.Println(err)
		res := page.Result{Result: "Error", ErrorDescription: "Error parsing form"}
		respBody, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			makeResponse(respBody, w, r)
			return
		}
	}
	name := r.Form.Get("name")
	img := r.Form.Get("img")
	text := r.Form.Get("text")
	if name == "" || img == "" || text == "" {
		res := page.Result{Result: "Error", ErrorDescription: "One or more parameter is empty"}
		respBody, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			makeResponse(respBody, w, r)
			return
		}
	}
	s.pages[pageId].Name = name
	s.pages[pageId].Img = img
	s.pages[pageId].Text = text
	res, err := json.Marshal(s.pages[pageId])
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	makeResponse(res, w, r)
}
func (s*Server) Add(w http.ResponseWriter, r*http.Request){
	s.Mu.Lock()
	defer s.Mu.Unlock()
	err := r.ParseForm()
	if err != nil{
		log.Println(err)
		res := page.Result{Result: "Error", ErrorDescription: "Error parsing form"}
		respBody, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			makeResponse(respBody, w, r)
			return
		}
	}

	name := r.PostForm.Get("name")
	img := r.PostForm.Get("img")
	text := r.PostForm.Get("text")
	if name == "" || img == "" || text == "" {
		res := page.Result{Result: "Error", ErrorDescription: "One or more parameter is empty"}
		respBody, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			makeResponse(respBody, w, r)
			return
		}
	}
	s.maxPageID += 1
	newPage := page.Page{Id: s.maxPageID, Name: name, Img: img, Text: text, Date: time.Now()}
	s.pages = append(s.pages, &newPage)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(CHANGED)
	res, err := json.Marshal(newPage)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	makeResponse(res, w, r)
}
func (s*Server) Delete(w http.ResponseWriter, r*http.Request){
	s.Mu.Lock()
	defer s.Mu.Unlock()
	found := false
	params, err := remux.PathParams(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	id, err := strconv.ParseInt(params.Named["Id"], 10, 64)
	if err != nil{
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	var pageId int64
	for i, pg := range s.pages{
		if id == pg.Id {
			pageId = int64(i)
			found = true
			break
		}
	}
	if !found{
		res := page.Result{Result: "Error", ErrorDescription: "Not found"}
		respBody, err := json.Marshal(res)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		makeResponse(respBody, w, r)
		return
	}
	if len(s.pages) == 1{
		s.pages = []*page.Page{}
		w.WriteHeader(DELETED)
		return
	}
	before := s.pages[:pageId]
	after := s.pages[pageId+1:]
	s.pages = before
	s.pages = append(s.pages, after...)
	w.WriteHeader(DELETED)
	return
}
func makeResponse(respBody []byte, w http.ResponseWriter, r*http.Request) {
	w.Header().Add("Content-Type", "application/json")
	_, err := w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}
