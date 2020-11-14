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
	OK      = 200
)

type Server struct {
	Mu        sync.RWMutex
	Mux       *remux.ReMUX
	pages     []*page.Page
	maxPageID int64
}

func NewService() *Server {
	mu := remux.CreateNewReMUX()

	return &Server{Mu: sync.RWMutex{}, Mux: mu, pages: []*page.Page{}, maxPageID: 0}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Mux.ServeHTTP(w, r)
}
func (s *Server) GetAll(w http.ResponseWriter, r *http.Request) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	pages := make([]page.PageDTO, 0, len(s.pages))
	for _, currPage := range s.pages {
		pg := page.PageDTO{Id: currPage.Id, Name: currPage.Name, Img: currPage.Img, Date: currPage.Date}
		pages = append(pages, pg)
	}

	err := makeResponse(pages, w, OK)
	if err != nil {
		log.Println(err)
		return
	}
}

func (s *Server) GetSingle(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	for _, pg := range s.pages {
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
	if !found {
		res := page.Result{Result: "Error", ErrorDescription: "Not found"}

		err := makeResponse(res, w, OK)
		if err != nil {
			log.Println(err)
		}
		return
	}

	err = makeResponse(result, w, OK)
	if err != nil {
		log.Println(err)
	}
	return
}
func (s *Server) Change(w http.ResponseWriter, r *http.Request) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	found := false
	params, err := remux.PathParams(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	id, err := strconv.ParseInt(params.Named["Id"], 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	var pageId int64
	for i, pg := range s.pages {
		if id == pg.Id {
			pageId = int64(i)
			found = true
			break
		}
	}
	if !found {
		res := page.Result{Result: "Error", ErrorDescription: "Not found"}
		err := makeResponse(res, w, OK)
		if err != nil {
			log.Println(err)
		}
		return

	}
	err = r.ParseForm()
	if err != nil {
		log.Println(err)
		res := page.Result{Result: "Error", ErrorDescription: "Error parsing form"}
		err := makeResponse(res, w, OK)
		if err != nil {
			log.Println(err)
		}
		return
	}
	name := r.Form.Get("name")
	img := r.Form.Get("img")
	text := r.Form.Get("text")
	if name == "" || img == "" || text == "" {
		res := page.Result{Result: "Error", ErrorDescription: "One or more parameter is empty"}
		err := makeResponse(res, w, OK)
		if err != nil {
			log.Println(err)
		}
		return
	}
	s.pages[pageId].Name = name
	s.pages[pageId].Img = img
	s.pages[pageId].Text = text
	err = makeResponse(s.pages[pageId], w, OK)
	if err != nil {
		log.Println(err)
		return
	}
}
func (s *Server) Add(w http.ResponseWriter, r *http.Request) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		res := page.Result{Result: "Error", ErrorDescription: "Error parsing form"}
		err := makeResponse(res, w, OK)
		if err != nil {
			log.Println(err)
		}
		return
	}

	name := r.PostForm.Get("name")
	img := r.PostForm.Get("img")
	text := r.PostForm.Get("text")
	if name == "" || img == "" || text == "" {
		res := page.Result{Result: "Error", ErrorDescription: "One or more parameter is empty"}
		err := makeResponse(res, w, OK)
		if err != nil {
			log.Println(err)
		}
		return
	}
	s.maxPageID += 1
	newPage := page.Page{Id: s.maxPageID, Name: name, Img: img, Text: text, Date: time.Now()}
	s.pages = append(s.pages, &newPage)
	err = makeResponse(newPage, w, CHANGED)
	if err != nil {
		log.Println(err)
		return
	}
}
func (s *Server) Delete(w http.ResponseWriter, r *http.Request) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	found := false
	params, err := remux.PathParams(r.Context())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	id, err := strconv.ParseInt(params.Named["Id"], 10, 64)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	var pageId int64
	for i, pg := range s.pages {
		if id == pg.Id {
			pageId = int64(i)
			found = true
			break
		}
	}
	if !found {
		res := page.Result{Result: "Error", ErrorDescription: "Not found"}
		err := makeResponse(res, w, OK)
		if err != nil {
			log.Println(err)
		}
		return

	}
	if len(s.pages) == 1 {
		s.pages = []*page.Page{}
		w.WriteHeader(DELETED)
		return
	}
	before := s.pages[:pageId]
	after := s.pages[pageId+1:]
	s.pages = before
	s.pages = append(s.pages, after...)
	result := page.Result{Result: "Deleted"}
	err = makeResponse(result, w, DELETED)
	if err != nil {
		log.Println(err)
		return
	}
}
func makeResponse(resp interface{}, w http.ResponseWriter, status int) error {
	if resp != nil {
		body, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		w.Header().Add("Content-Type", "application/json")
		_, err = w.Write(body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
	}
	w.WriteHeader(status)
	return nil
}
