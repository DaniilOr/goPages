package page

import "time"

type Page struct{
	Id int64
	Name string
	Img string
	Text string
	Date time.Time
}

type PageDTO struct{
	Id int64
	Name string
	Img string
	Date time.Time
}

type Result struct{
	Result string `json:"result"`
	ErrorDescription string `json:"errorDesc,omitempty"`

}