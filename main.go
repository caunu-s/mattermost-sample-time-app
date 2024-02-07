package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

//go:embed icon.png
var IconData []byte

var Manifest = apps.Manifest{
	AppID: "datetime",
	Version: "v0.1.0",
	DisplayName: "Datetime All over The World",
	Icon: "icon.png",
	HomepageURL: "https://github.com/caunu-s/mattermost-sample-time-app",
	RequestedPermissions: []apps.Permission{
		apps.PermissionActAsBot,
		apps.PermissionActAsUser,
	},
	RequestedLocations: []apps.Location{
		apps.LocationChannelHeader,
		apps.LocationCommand,
	},
	Deploy: apps.Deploy{
		HTTP: &apps.HTTP{
			RootURL: "http://mattermost-apps-datetime:4000",
		},
	},
}

var Bindings = []apps.Binding{
	{
		Location: apps.LocationChannelHeader,
		Bindings: []apps.Binding{
			{
				Location: "send-button",
				Icon:     "icon.png",
				Label:    "check datetime",
				Form:     &SendForm,
			},
		},
	},
	{
		Location: "/command",
		Bindings: []apps.Binding{
			{
				Icon:        "icon.png",
				Label:       "datetime",
				Description: "Datetime app",
				Hint:        "[send]",
				Bindings: []apps.Binding{
					{
						Label: "send",
						Form:  &SendForm,
					},
				},
			},
		},
	},
}

var SendForm = apps.Form{
	Title: "Datetime All over The World",
	Icon:  "icon.png",
	Fields: []apps.Field{
		{
			Type: "text",
			Name: "Timezone",
		},
	},
	Submit: apps.NewCall("/send").WithExpand(apps.Expand{
		ActingUserAccessToken: apps.ExpandAll,
		ActingUser:            apps.ExpandID,
	}),
}

type Response struct {
	Datetime string `json:"datetime"`
}

func main() {
	http.HandleFunc("/manifest.json",
		httputils.DoHandleJSON(Manifest))
	http.HandleFunc("/static/icon.png",
		httputils.DoHandleData("image/png", IconData))
	http.HandleFunc("/bindings",
		httputils.DoHandleJSON(apps.NewDataResponse(Bindings)))
	http.HandleFunc("/send", Send)

	addr := ":4000"
	fmt.Println("Listening on", addr)
	fmt.Println("Use '/apps install http http://mattermost-apps-datetime" + addr + "/manifest.json' to install the app")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func Send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	url := "http://worldtimeapi.org/api/timezone/"
	v, ok := c.Values["Timezone"]
	if ok && v != nil {
		url += fmt.Sprintf("%s", v)
	} else {
		url += "Asia/Tokyo"
	}

	res_time, _ := http.Get(url)
	defer res_time.Body.Close()
	rbody, _ := ioutil.ReadAll(res_time.Body)
	var response Response
	json.Unmarshal(rbody, &response)

	httputils.WriteJSON(w,
		apps.NewTextResponse(string(response.Datetime)))
}