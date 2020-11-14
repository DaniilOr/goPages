package page

import "time"

type Page struct {
	Id   int64     `json:"id"`
	Name string    `json:"name"`
	Img  string    `json:"img"`
	Text string    `json:"text"`
	Date time.Time `json:"date"`
}

type PageDTO struct {
	Id   int64
	Name string
	Img  string
	Date time.Time
}

type Result struct {
	Result           string `json:"result"`
	ErrorDescription string `json:"errorDesc,omitempty"`
}
