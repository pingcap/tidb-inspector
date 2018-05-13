package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/unrolled/render"
)

// KV is a set of key/value string pairs
type KV map[string]string

// Alert holds one alert for notification templates
type Alert struct {
	Status       string    `json:"status"`
	Labels       KV        `json:"labels"`
	Annotations  KV        `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
}

// Alerts is a list of Alert objects
type Alerts []Alert

// AlertData is base structure of alertmanager
type AlertData struct {
	Receiver          string `json:"receiver"`
	Status            string `json:"status"`
	Alerts            Alerts `json:"alerts"`
	GroupLabels       KV     `json:"groupLabels"`
	CommonLabels      KV     `json:"commonLabels"`
	CommonAnnotations KV     `json:"commonAnnotations"`
	ExternalURL       string `json:"externalURL"`
}

func readJSON(r io.ReadCloser, data interface{}) error {
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err == nil {
		err = json.Unmarshal(b, data)
	}

	return errors.Trace(err)
}

// AlertMsgFromWebhook get alert massage
func (r *Run) AlertMsgFromWebhook(w http.ResponseWriter, hr *http.Request) {
	alertData := &AlertData{}
	err := readJSON(hr.Body, alertData)
	if err != nil {
		log.Errorf("can not unmarshal http post data with error %v", err)
		r.Rdr.Text(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Infof("alert data %+v", alertData)

	r.AlertMsgs <- alertData
	r.Rdr.Text(w, http.StatusAccepted, "success")
}

// CreateRouter creates router
func (r *Run) CreateRouter() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc("/v1/alertmanager", r.AlertMsgFromWebhook).Methods("POST")

	return m
}

// CreateRender creates a new Render
func (r *Run) CreateRender() {
	r.Rdr = render.New(render.Options{
		IndentJSON: true,
	})
}
