package main

import (
	"errors"
	"fmt"
	"github.com/pingcap/kvproto/pkg/pdpb"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	//healthAPI get pd cluster status
	healthAPI = "/pd/health"
	//pingAPI check pd client status
	pingAPI = "/pd/ping"
	//oldHealthAPI get etcd cluster status
	oldHealthAPI = "/health"
	//membersAPI get pd members status
	membersAPI = "/pd/api/v1/members"
)

type pdHealth struct {
	Name       string   `json:"name"`
	MemberID   uint64   `json:"member_id"`
	ClientUrls []string `json:"client_urls"`
	Health     bool     `json:"health"`
}

type pdHealths []pdHealth

/*
{
  "members": [
    {
      "name": "pd1",
      "member_id": 7423789596866920899,
      "peer_urls": [
        "http://10.0.3.4:2380"
      ],
      "client_urls": [
        "http://10.0.3.4:2379"
      ]
    },
  ]
}
*/
func getMembers(pdURL string) ([]*pdpb.Member, error) {
	var pdMembers map[string][]*pdpb.Member

	err := xGet(pdURL, &pdMembers, true)
	//pdMembers["members"] is hardcode
	log.Debugf("get memebers %v error %v", pdMembers, err)
	return pdMembers["members"], err
}

/*
oldHealthAPI only return etcd cluster status
{"health": "true"}

healthAPI return all node status
[
  {
    "name": "pd1",
    "member_id": 7423789596866920899,
    "client_urls": ["http://10.0.3.4:2379"],
    "health": true
  },
]
*/
func checkPD(pdURLs []string) (pdHealths, error) {
	var pdhs pdHealths
	var err error
	// new api
	//log.Infof("check pd health with new api")
	for _, pdURL := range pdURLs {
		newPDHealthURL := fmt.Sprintf("%s%s", pdURL, healthAPI)
		if err = xGet(newPDHealthURL, &pdhs, true); err == nil {
			break
		}
		log.Errorf("new api get healths failed from url %s with error %v", newPDHealthURL, err)
	}
	if len(pdhs) > 0 {
		return pdhs, nil
	}

	//old api
	//log.Infof("check pd health with old api")
	var pdMembers []*pdpb.Member
	for _, pdURL := range pdURLs {
		pdMembersURL := fmt.Sprintf("%s%s", pdURL, membersAPI)
		if pdMembers, err = getMembers(pdMembersURL); err == nil {
			break
		}
	}
	if len(pdMembers) == 0 {
		return pdhs, errors.New("can not get memebers for pd url")
	}
	for _, member := range pdMembers {
		h := pdHealth{
			Name:       member.Name,
			MemberID:   member.MemberId,
			ClientUrls: member.ClientUrls,
			Health:     true,
		}
		for _, cURL := range member.ClientUrls {
			pingURL := fmt.Sprintf("%s%s", cURL, pingAPI)
			if err := xGet(pingURL, nil, false); err != nil {
				h.Health = false
			}
		}
		pdhs = append(pdhs, h)
	}
	return pdhs, nil
}

func goroutinePD(pdURLs []string, checkInterval time.Duration) {
	for {
		if pdhs, err := checkPD(pdURLs); err != nil {
			log.Debugf("check pd failed with error %v", err)
			exporter.WithLabelValues(promPDType, checkedAllInstance, checkedFailed).Inc()
		} else {
			for _, pdh := range pdhs {
				checked := checkedSuccess
				if !pdh.Health {
					checked = checkedFailed
					xAlert("", promPDType, pdh.Name, checked)
				}
				exporter.WithLabelValues(promPDType, pdh.Name, checked).Inc()
			}
		}
		time.Sleep(checkInterval)
	}

}
