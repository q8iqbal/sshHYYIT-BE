package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Log struct {
	Id        primitive.ObjectID `json:"id,omitempty"`
	Ip_Server string             `json:"ip_server,omitempty" validate:"required"` // ip dari server vps
	Hostname  string             `json:"hostname,omitempty" validate:"required"`  // nama host vps
	Ip_Guest  string             `json:"ip_guest,omitempty" validate:"required"`  // ip dari client yang berusaha connect
	Username  string             `json:"username,omitempty" validate:"required"`  // nama dari user yang digunakan
	Timestamp string             `json:"timestamp,omitempty" validate:"required"` // tanggal waktu log
	District  string             `json:"district,omitempty"`                      // diproses di back end
	State     string             `json:"state,omitempty"`                         // diproses di back end
	Country   string             `json:"country,omitempty"`                       // diproses di back end
	Status    string             `json:"status,omitempty" validate:"required"`    // status log, connected atau failed
}

type GeoIP struct {
	Ip           string `json:"ip"`
	State_Prov   string `json:"state_prov"`
	District     string `json:"district"`
	Country_name string `json:"country_name"`
}
type CurrentServer struct {
	Ip_Server string        `json:"ip_server"`
	Hostname  string        `json:"hostname"`
	Users     []CurrentUser `json:"users"`
}
type CurrentUser struct {
	User      string `json:"user,omitempty" validate:"required"`
	Ip_Guest  string `json:"ip_guest,omitempty" validate:"required"`  // ip dari client yang berusaha connect
	Timestamp string `json:"timestamp,omitempty" validate:"required"` // tanggal waktu log
}

type LoginUser struct {
	Username string `json:"username,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"` // ip dari client yang berusaha connect
}
