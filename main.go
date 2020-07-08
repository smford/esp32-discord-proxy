package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	Token string
	Color = 0x009688
	//Icons = "https://kittyhacker101.tk/Static/KatBot"
	Icons    = "https://cdn.discordapp.com/emojis"
	Emojis   = make(map[string]string)
	Channels = make(map[string]string)
	//Devices   = make(map[string]string)
	DeviceMap = make(map[string]deviceStruct)
	ActionMap = make(map[string]actionStruct)
)

const APITOKEN = "sometoken"
const DEVICEAPITOKEN = "xyz"
const LISTENIP = "0.0.0.0"
const LISTENPORT = "57000"
const INDEXHTML = "index.html"

type wifiNetwork struct {
	RSSI    int    `json:"rssi"`
	SSID    string `json:"ssid"`
	BSSID   string `json:"bssid"`
	Channel int    `json:"channel"`
	Secure  int    `json:"secure"`
}

type shortStatusResult struct {
	Timestamp       string `json:"Timestamp"`
	Hostname        string `json:"Hostname"`
	Device          string `json:"Device"`
	State           string `json:"State"`
	Maintenancemode string `json:"MaintenanceMode"`
}

type deviceStruct struct {
	Name     string `json:"Name"`
	Hostname string `json:"Hostname"`
	Port     int    `json:"Port"`
	Channel  string `json:"Channel"`
}

type actionStruct struct {
	Name   string `json:"Name"`
	States []string
}

//var MyDevices []deviceStruct

func init() {
	Emojis["laseron"] = "<:laseron:729726642758615151>"
	Emojis["laseroff"] = "<:laseroff:730213748102529064>"
	Emojis["maintenanceon"] = "<:maintenanceon:729732695009263616>"
	Emojis["maintenanceoff"] = "<:maintenanceoff:729828147414958202>"
	Emojis["backlighton"] = "<:backlighton:729820542336761856>"
	Emojis["backlightoff"] = "<:backlightoff:729820688516644894>"
	Emojis["eehtick"] = "<:eehtick:729828147414958202>"
	Emojis["overrideon"] = "<:overrideon:730075631198404649>"
	Emojis["overrideoff"] = "<:overrideoff:730448103517454376>"
	Emojis["3don"] = "<:3don:730213748102529064>"
	Emojis["3doff"] = "<3doff:730213748102529064>"
	Emojis["userlogin"] = "<:userlogin:730444250839515297>"
	Emojis["userlogout"] = "<:userlogout:730444251695153285>"

	Channels["general-junk"] = "729631967905054764"
	Channels["laser"] = "729632142358872138"

	// Devices["laser"] = "192.168.10.135"

	/* MyDevices = []deviceStruct{
		deviceStruct{
			Name:    "laser",
			IP:      "192.168.10.135",
			Channel: "729632142358872138", // laser
		},
		deviceStruct{
			Name:    "laser2",
			IP:      "192.168.10.136",
			Channel: "729632142358872138", // laser
		},
	} */

	DeviceMap["laser"] = deviceStruct{
		Name:     "laser",
		Hostname: "192.168.10.135",
		Port:     80,
		Channel:  "729632142358872138", // laser
	}

	DeviceMap["laser2"] = deviceStruct{
		Name:     "laser",
		Hostname: "192.168.10.136",
		Port:     80,
		Channel:  "729632142358872138", // laser
	}

	ActionMap["maintenance"] = actionStruct{
		Name:   "maintenance",
		States: []string{"on", "off"},
	}

	ActionMap["backlight"] = actionStruct{
		Name:   "backlight",
		States: []string{"on", "off"},
	}

	ActionMap["override"] = actionStruct{
		Name:   "override",
		States: []string{"on", "off"},
	}

	ActionMap["user"] = actionStruct{
		Name:   "user",
		States: []string{"login", "logout"},
	}

	/*
		var steve string = "laser"
		fmt.Println("Name:", DeviceMap[steve].Name, " Host:", DeviceMap[steve].Hostname, ":", DeviceMap[steve].Port, " Channel:", DeviceMap[steve].Channel)
		steve = "laser2"
		fmt.Println("Name:", DeviceMap[steve].Name, " Host:", DeviceMap[steve].Hostname, ":", DeviceMap[steve].Port, " Channel:", DeviceMap[steve].Channel)
	*/

	for k, v := range DeviceMap {
		fmt.Printf("%s  Host:%s:%d  Channel:%s\n", k, v.Hostname, v.Port, v.Channel)
	}
	//os.Exit(0)

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	startWeb(LISTENIP, LISTENPORT, false)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!laser backlight on" {
		s.ChannelMessageSend(m.ChannelID, backlight("on"))
	}

	if m.Content == "!laser backlight off" {
		s.ChannelMessageSend(m.ChannelID, backlight("off"))
	}

	if m.Content == "!laser fullstatus" {
		s.ChannelMessageSend(m.ChannelID, fullStatus())
	}

	if m.Content == "!laser help" {
		var printText string
		printText += "```\n"
		printText += "Available Commands:\n"
		printText += "-------------------------------\n"
		printText += "  laser backlight [on|off]\n"
		printText += "  laser fullstatus\n"
		printText += "  laser help\n"
		printText += "  laser maintenance [enable|disable]\n"
		printText += "  laser scanwifi\n"
		printText += "  laser status\n"
		printText += "```\n"
		s.ChannelMessageSend(m.ChannelID, printText)
	}

	if m.Content == "!laser maintenance disable" {
		s.ChannelMessageSend(m.ChannelID, maintenancemode("disable"))
	}

	if m.Content == "!laser maintenance enable" {
		s.ChannelMessageSend(m.ChannelID, maintenancemode("enable"))
	}

	if m.Content == "!laser scanwifi" {
		s.ChannelMessageSend(m.ChannelID, scanWifi())
	}

	if m.Content == "!laser status" {
		s.ChannelMessageSend(m.ChannelID, shortStatus())
	}

	if m.Content == "!cat" {
		tr := &http.Transport{DisableKeepAlives: true}
		client := &http.Client{Transport: tr}
		resp, err := client.Get("https://images-na.ssl-images-amazon.com/images/I/71FcdrSeKlL._AC_SL1001_.jpg")
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Unable to fetch cat!")
			fmt.Println("[Warning] : Cat API Error")
		} else {
			s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{Name: "Cat Picture", IconURL: Icons + "/729726642758615151.png"},
				Color:  Color,
				Image: &discordgo.MessageEmbedImage{
					URL: resp.Request.URL.String(),
				},
				Footer: &discordgo.MessageEmbedFooter{Text: "Cat pictures provided by TheCatApi", IconURL: Icons + "/729726642758615151.png"},
			})
			fmt.Println("[Info] : Cat sent successfully to " + m.Author.Username + "(" + m.Author.ID + ") in " + m.ChannelID)
		}
	}
}

func shortStatus() string {
	fmt.Println("starting shortStatus")
	url := fmt.Sprintf("http://192.168.10.135/status")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return ""
	}
	client := &http.Client{}
	client.Timeout = time.Second * 15
	resp, err := client.Do(req)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var mystatus shortStatusResult
		err = json.Unmarshal([]byte(bodyBytes), &mystatus)
		if err != nil {
			fmt.Println("unmarshal error: ")
			fmt.Println(err)
		}
		var returnText = ""
		if mystatus.State == "on" {
			returnText = Emojis["laseron"] + " **" + strings.ToUpper(mystatus.Device) + " IN USE**"
		} else {
			returnText = "**" + strings.ToUpper(mystatus.Device) + " IS FREE**"
		}
		if mystatus.Maintenancemode == "enabled" {
			returnText = Emojis["maintenance"] + " **" + strings.ToUpper(mystatus.Device) + " IN MAINTENANCE MODE**"
		}
		return returnText
	}
	return "No status available"
}

func fullStatus() string {
	fmt.Println("starting fullStatus")
	url := fmt.Sprintf("http://192.168.10.135/fullstatus?api=%s", DEVICEAPITOKEN)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return ""
	}
	client := &http.Client{}
	client.Timeout = time.Second * 15
	resp, err := client.Do(req)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var mystatus shortStatusResult
		err = json.Unmarshal([]byte(bodyBytes), &mystatus)
		if err != nil {
			fmt.Println("unmarshal error: ")
			fmt.Println(err)
		}
		var returnText = ""
		if mystatus.State == "on" {
			returnText = Emojis["laseron"] + " **" + strings.ToUpper(mystatus.Device) + " IN USE**"
		} else {
			returnText = "**" + strings.ToUpper(mystatus.Device) + " IS FREE**"
		}
		if mystatus.Maintenancemode == "enabled" {
			returnText = Emojis["maintenance"] + " **" + strings.ToUpper(mystatus.Device) + " IN MAINTENANCE MODE**"
		}
		return returnText
	}
	return "No status available"
}

func scanWifi() string {
	fmt.Println("starting scanWifi")
	url := fmt.Sprintf("http://192.168.10.135/scanwifi?api=%s", DEVICEAPITOKEN)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return ""
	}
	client := &http.Client{}
	client.Timeout = time.Second * 15
	resp, err := client.Do(req)
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var mynetworks []wifiNetwork
		err = json.Unmarshal([]byte(bodyBytes), &mynetworks)
		if err != nil {
			fmt.Println("unmarshal error: ")
			fmt.Println(err)
		}
		var returnText = ""

		fmt.Println("Number of networks found: ", len(mynetworks))

		if len(mynetworks) > 0 {
			returnText = "```\n"
		}

		for i := 0; i < len(mynetworks); i++ {
			fmt.Println(mynetworks[i].SSID)
			returnText += mynetworks[i].SSID + "\n"
		}
		if len(mynetworks) > 0 {
			returnText += "```"
		}
		fmt.Println(returnText)
		return returnText
	}
	return "No networks available"
}

func backlightold(mystate string) string {
	fmt.Println("starting backlight")
	url := fmt.Sprintf("http://192.168.10.135/backlight?api=%s&state=%s", DEVICEAPITOKEN, mystate)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return "ERROR"
	}
	client := &http.Client{}
	client.Timeout = time.Second * 15
	resp, err := client.Do(req)
	if resp.StatusCode == http.StatusOK {
		if err != nil {
			log.Fatal(err)
		}
		if mystate == "on" {
			return Emojis["backlighton"] + " **LASER BACKLIGHT ON**"
		}
		if mystate == "off" {
			return Emojis["backlightoff"] + " **LASER BACKLIGHT OFF**"
		}
	} else {
		return "ERROR"
	}
	return ""
}

//=========
func backlight(mystate string) string {
	fmt.Println("starting backlight")
	url := fmt.Sprintf("http://192.168.10.135/backlight?api=%s&state=%s", DEVICEAPITOKEN, mystate)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return "ERROR"
	}
	client := &http.Client{}
	client.Timeout = time.Second * 15
	client.Do(req)
	return ""
}

//=========

func maintenancemode(mystate string) string {
	fmt.Println("starting maintenancemode")
	url := fmt.Sprintf("http://192.168.10.135/maintenance?api=%s&state=%s", DEVICEAPITOKEN, mystate)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		return "ERROR"
	}
	client := &http.Client{}
	client.Timeout = time.Second * 15
	//resp, err := client.Do(req)
	client.Do(req)
	/*if resp.StatusCode == http.StatusOK {
		if err != nil {
			log.Fatal(err)
		}
		if mystate == "enable" {
			return Emojis["maintenance"] + " **LASER IN MAINTENANCE MODE**"
		}
		if mystate == "disable" {
			return Emojis["eehtick"] + " **LASER IS AVAILABLE TO USE**"
		}
	} else {
		return "ERROR"
	}*/
	return ""
}

func printFile(filename string, webprint http.ResponseWriter) {
	fmt.Println("Starting printFile")
	texttoprint, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("ERROR: cannot open ", filename)
		if webprint != nil {
			http.Error(webprint, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}
	if webprint != nil {
		fmt.Fprintf(webprint, "%s", string(texttoprint))
	} else {
		fmt.Print(string(texttoprint))
	}
}

func startWeb(listenip string, listenport string, usetls bool) {
	r := mux.NewRouter()

	r.HandleFunc("/", handlerIndex)

	r.HandleFunc("/laser", handlerLaser)
	//laserRouter := r.PathPrefix("/laser").Subrouter()
	//laserRouter.HandleFunc("/{laser}", handlerLaser)
	//laserRouter.Use(loggingMiddleware)

	r.HandleFunc("/api", handlerApi)

	log.Printf("Starting HTTP Webserver http://%s:%s\n", listenip, listenport)

	srv := &http.Server{
		Handler:      r,
		Addr:         LISTENIP + ":" + LISTENPORT,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err := srv.ListenAndServe()

	fmt.Println("cannot start http server:", err)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println("MIDDLEWARE: ", r.RemoteAddr, " ", r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting handlerIndex")
	printFile(INDEXHTML, w)
}

func handlerLaser(webprint http.ResponseWriter, r *http.Request) {
	fmt.Println("starting handlerLaser2")
	queries := r.URL.Query()
	fmt.Printf("queries = %q\n", queries)

	if APITOKEN != queries.Get("api") {
		fmt.Fprintf(webprint, "%s", "ERROR: Invalid API")
		return
	}

	var returnText = ""

	switch strings.ToLower(queries.Get("action")) {
	case "off":
		returnText = Emojis["eehtick"] + " **LASER IS AVAILABLE TO USE**"
	case "on":
		returnText = Emojis["laseron"] + " **" + queries.Get("user") + " IS FIRING LASER, PEW PEW**"
	case "override":
		returnText = Emojis["eehboss"] + " **LASER BOSS MODE ENABLED**"
	case "maintenanceon":
		returnText = Emojis["maintenance"] + " **LASER IN MAINTENANCE MODE**"
	default:
		return
	}

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Unable to create discord session!")
		return
	}

	fmt.Fprintf(webprint, "%s", returnText)
	dg.ChannelMessageSend("729631967905054764", returnText)

}

//====

func checkIfDeviceExists(device string) bool {
	// get all the device names and store in an array
	i := 0
	devicearray := make([]string, len(DeviceMap))
	for k := range DeviceMap {
		devicearray[i] = k
		i++
	}

	for _, a := range devicearray {
		if strings.ToLower(a) == strings.ToLower(device) {
			return true
		}
	}
	return false
}

// check if valid action and state
func checkActionAndState(action string, state string) bool {
	// get all the device names and store in an array
	i := 0
	actionarray := make([]string, len(ActionMap))
	for k := range ActionMap {
		actionarray[i] = k
		i++
	}

	//fmt.Printf("array=%v\n", actionarray)

	for _, a := range actionarray {
		//fmt.Println("a=", a)
		if strings.ToLower(a) == strings.ToLower(action) {
			//fmt.Println("Valid Action:", action)
			//fmt.Println("strings.ToLower(a)=", strings.ToLower(a), "  strings.ToLower(action)=", strings.ToLower(action))
			for _, s := range ActionMap[a].States {
				//fmt.Println("s=", s)
				//fmt.Printf("v=%v\n", s)
				if strings.ToLower(s) == strings.ToLower(state) {
					// valid state for this action found
					return true
				}
			}
		}
	}
	// if reached here, state or action is bad and thus return false
	return false
}

func handlerApi(webprint http.ResponseWriter, r *http.Request) {
	fmt.Println("starting handlerLaser2")
	queries := r.URL.Query()
	fmt.Printf("queries = %q\n", queries)

	// check if api token is valid
	if APITOKEN != queries.Get("token") {
		fmt.Println("ERROR: Invalid API Token", queries.Get("token"))
		fmt.Fprintf(webprint, "%s", "ERROR: Invalid API Token")
		return
	}

	if !checkIfDeviceExists(queries.Get("device")) {
		fmt.Println("ERROR: Invalid device", queries.Get("device"))
		fmt.Fprintf(webprint, "%s", "ERROR: Invalid Device")
		return
	}

	if !checkActionAndState(queries.Get("action"), queries.Get("state")) {
		fmt.Println("ERROR: Bad action or state", queries.Get("action"), queries.Get("state"))
		fmt.Fprintf(webprint, "%s", "ERROR: Bad action or state")
		return
	}

	fmt.Printf("Device %s is valid\nAction %s is valid\nState %s is valid\n", queries.Get("device"), queries.Get("action"), queries.Get("state"))
	fmt.Fprintf(webprint, "Device %s is valid\nAction %s is valid\nState %s is valid", queries.Get("device"), queries.Get("action"), queries.Get("state"))

	var returnText = ""

	/* switch strings.ToLower(queries.Get("action")) {
	case "off":
		returnText = Emojis["eehtick"] + " **LASER IS AVAILABLE TO USE**"
	case "on":
		returnText = Emojis["laseron"] + " **" + queries.Get("user") + " IS FIRING LASER, PEW PEW**"
	case "override":
		returnText = Emojis["eehboss"] + " **LASER BOSS MODE ENABLED**"
	case "maintenanceon":
		returnText = Emojis["maintenance"] + " **LASER IN MAINTENANCE MODE**"
	default:
		return
	}
	*/

	var lookup = queries.Get("action") + queries.Get("state")
	fmt.Println("lookup=", lookup)
	returnText = Emojis[queries.Get("action")+queries.Get("state")] + " **" + strings.ToUpper(queries.Get("device")) + " " + strings.ToUpper(queries.Get("action")) + ":" + strings.ToUpper(queries.Get("state")) + "**"

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Unable to create discord session!")
		return
	}

	//fmt.Fprintf(webprint, "%s,  %s", DeviceMap[queries.Get("device")].Channel, returnText)
	dg.ChannelMessageSend(DeviceMap[queries.Get("device")].Channel, returnText)
	fmt.Println("returnText = ", returnText)
	//dg.ChannelMessageSend("729631967905054764", returnText)

}
