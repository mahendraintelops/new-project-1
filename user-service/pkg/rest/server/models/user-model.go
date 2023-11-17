package models

type User struct {
	Id int64 `json:"id,omitempty"`

	City string `json:"city,omitempty"`

	Name string `json:"name,omitempty"`
}
