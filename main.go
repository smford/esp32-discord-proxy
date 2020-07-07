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
	Icons = "https://cdn.discordapp.com/emojis"
)

const APITOKEN = "xyz"
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

func init() {
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
		//s.ChannelMessageSend(s.UserChannelCreate(m.Author.ID), printText)
		//privatechan, err := s.UserChannelCreate(m.Author.ID)

		//if err != nil {
		//	fmt.Println("ERROR: ", err)
		//}

		//fmt.Printf("privatechan:%-v", privatechan)
		//privatechan.ChannelMessageSend(printText)
		//var privatechan = discordgo.Session
		//privatechan.ChannelMessageSend(privatechanid, printText)
		//s.UserChannelCreate(m.Author.ID,
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
			returnText = "<:laseron:729726642758615151> **" + strings.ToUpper(mystatus.Device) + " IN USE**"
		} else {
			returnText = "**" + strings.ToUpper(mystatus.Device) + " IS FREE**"
		}
		if mystatus.Maintenancemode == "enabled" {
			returnText = "<:lasermaintenance:729732695009263616> **" + strings.ToUpper(mystatus.Device) + " IN MAINTENANCE MODE**"
		}
		return returnText
	}
	return "No status available"
}

func fullStatus() string {
	fmt.Println("starting fullStatus")
	return "full status disabled"
	url := fmt.Sprintf("http://192.168.10.135/fullstatus?api=%s", APITOKEN)
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
			returnText = "<:laseron:729726642758615151> **" + strings.ToUpper(mystatus.Device) + " IN USE**"
		} else {
			returnText = "**" + strings.ToUpper(mystatus.Device) + " IS FREE**"
		}
		if mystatus.Maintenancemode == "enabled" {
			returnText = "<:lasermaintenance:729732695009263616> **" + strings.ToUpper(mystatus.Device) + " IN MAINTENANCE MODE**"
		}
		return returnText
	}
	return "No status available"
}

func scanWifi() string {
	fmt.Println("starting scanWifi")
	url := fmt.Sprintf("http://192.168.10.135/scanwifi?api=%s", APITOKEN)
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

func backlight(mystate string) string {
	fmt.Println("starting backlight")
	url := fmt.Sprintf("http://192.168.10.135/backlight?api=%s&state=%s", APITOKEN, mystate)
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
			return "<:backlighton:729820542336761856> **LASER BACKLIGHT ON**"
		}
		if mystate == "off" {
			return "<:backlightoff:729820688516644894> **LASER BACKLIGHT OFF**"
		}
	} else {
		return "ERROR"
	}
	return ""
}

func maintenancemode(mystate string) string {
	fmt.Println("starting maintenancemode")
	url := fmt.Sprintf("http://192.168.10.135/maintenance?api=%s&state=%s", APITOKEN, mystate)
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
		if mystate == "enable" {
			return "<:lasermaintenance:729732695009263616> **LASER IN MAINTENANCE MODE**"
		}
		if mystate == "disable" {
			return "<:eehtick:729828147414958202> **LASER IN NORMAL MODE**"
		}
	} else {
		return "ERROR"
	}
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
	fmt.Println("starting handlerLaser")
	fmt.Fprintf(webprint, "%s", "some text")

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Unable to create discord session!")
		return
	}

	dg.ChannelMessageSend("729631967905054764", "laser fired")

}
