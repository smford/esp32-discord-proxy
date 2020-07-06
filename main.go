package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
)

const APITOKEN = "xyz"

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
