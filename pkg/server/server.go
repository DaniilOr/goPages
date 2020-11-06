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

	return &Server{ sync.RWMutex{}, mu, []*page.Page{}, 0}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Mux.ServeHTTP(w, r)
}
func (s*Server) GetAll(w http.ResponseWriter, r*http.Request){
	log.Println(s.pages)
	pages := make([]page.PageDTO, len(s.pages))
	for i, page := range s.pages{
		pages[i].Id = page.Id
		pages[i].Name = page.Name
		pages[i].Img = page.Img
		pages[i].Date = page.Date
	}
	log.Println(pages)
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
	var pageId int64
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
		res := page.Result{"Error", "Error parsing form"}
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
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	var pageId int64
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
		res := page.Result{"Error", "One or more parameter is empty"}
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
	log.Println(newPage)
	s.pages = append(s.pages, &newPage)
	w.WriteHeader(CHANGED)
	res, err := json.Marshal(s.pages[pageId])
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	makeResponse(res, w, r)
}
func (s*Server) Detele(w http.ResponseWriter, r*http.Request){
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	var pageId int64
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
	if len(s.pages) == 1{
		s.pages = []*page.Page{}
		w.WriteHeader(DELETED)
		return
	}
	before := s.pages[:pageId-1]
	after := s.pages[pageId:]
	s.pages = before
	s.pages = append(s.pages, after...)
	w.WriteHeader(DELETED)
	res := page.Result{Result: "Done", ErrorDescription: "Page deleted"}
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
func makeResponse(respBody []byte, w http.ResponseWriter, r*http.Request) {
	w.Header().Add("Content-Type", "application/json")
	_, err := w.Write(respBody)
	if err != nil {
		log.Println(err)
	}
}
