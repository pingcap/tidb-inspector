package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/ngaut/log"
	"github.com/unrolled/render"
	"io/ioutil"
	"net/http"
	"time"
)

// AlertData alertmanager base struct
type AlertData struct {
	Receiver          string `json:"receiver"`
	Status            string `json:"status"`
	Alerts            Alerts `json:"alerts"`
	GroupLabels       KV     `json:"groupLabels"`
	CommonLabels      KV     `json:"commonLabels"`
	CommonAnnotations KV     `json:"commonAnnotations"`
	ExternalURL       string `json:"externalURL"`
}

// Alert holds one alert for notification templates.
type Alert struct {
	Status       string    `json:"status"`
	Labels       KV        `json:"labels"`
	Annotations  KV        `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
}

// Alerts is a list of Alert objects.
type Alerts []Alert

// KV is a set of key/value string pairs.
type KV map[string]string

//AlertMsgFromWebhook get alert massage
func (r *Run) AlertMsgFromWebhook(w http.ResponseWriter, hr *http.Request) {
	b, err := ioutil.ReadAll(hr.Body)
	if err != nil {
		log.Errorf("can not read http post data with error %v", err)
		r.Rdr.Text(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Debugf("get alert data b %v", string(b))
	defer hr.Body.Close()
	alertData := &AlertData{}
	err = json.Unmarshal(b, alertData)
	if err != nil {
		log.Errorf("can not unmarshal http post data with error %v", err)
		r.Rdr.Text(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Debugf("get alert data %v", alertData)
	r.AlertMsgs <- alertData
	r.Rdr.Text(w, http.StatusAccepted, "")
}

//CreateRouter create router
func (r *Run) CreateRouter() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/v1/alertmanager", r.AlertMsgFromWebhook).Methods("POST")

	return m
}

// CreateRender for render.
func (r *Run) CreateRender() {
	r.Rdr = render.New(render.Options{
		IndentJSON: true,
	})
}
