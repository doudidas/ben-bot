// Package p contains an HTTP Cloud Function.
package p

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// Set from Env Variables
var botToken string
var authorID string

type slackMessage struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}
type slackMessageResponse struct {
	Ok      bool   `json:"ok"`
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
	Message struct {
		Text        string `json:"text"`
		Username    string `json:"username"`
		BotID       string `json:"bot_id"`
		Attachments []struct {
			Text     string `json:"text"`
			ID       int    `json:"id"`
			Fallback string `json:"fallback"`
		} `json:"attachments"`
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
		Ts      string `json:"ts"`
	} `json:"message"`
}
type payload struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		ClientMsgID string `json:"client_msg_id"`
		Type        string `json:"type"`
		Text        string `json:"text"`
		User        string `json:"user"`
		Username    string `json:"username"`
		Ts          string `json:"ts"`
		Channel     string `json:"channel"`
		EventTs     string `json:"event_ts"`
		ChannelType string `json:"channel_type"`
	} `json:"event"`
	Type        string   `json:"type"`
	EventID     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
	AuthedUsers []string `json:"authed_users"`
}

// Joke from API
type Joke struct {
	Attachments []struct {
		Fallback string `json:"fallback"`
		Footer   string `json:"footer"`
		Text     string `json:"text"`
	} `json:"attachments"`
	ResponseType string `json:"response_type"`
	Username     string `json:"username"`
}

// Challenge from Slack
type Challenge struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

// ChallengeResponse response from Slack Challenge
type ChallengeResponse struct {
	Challenge string `json:"challenge"`
}

// Main function. Will manage message received from Slack
// and will send back a reply to the matching channel
func Main(w http.ResponseWriter, r *http.Request) {

	// Set Global Variables
	initEnv()
	var output, input []byte
	var c Challenge
	// dump(r)
	input = getInput(r)
	c = getChallenge(input)

	if c.Challenge != "" {
		output = replyChallenge(c)
	} else {
		fmt.Println("Not a Challenge")
		p := getPayload(input)
		go managePayload(p)
	}
	fmt.Println("output :>", output)
	w.Write(output)
}
func dump(r *http.Request) {
	byts, _ := httputil.DumpRequest(r, true)
	fmt.Println(string(byts))
}
func getInput(r *http.Request) []byte {
	var out []byte
	out, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error decoding input: " + err.Error())
	}
	fmt.Println("Input :> " + string(out))
	return out
}
func getChallenge(b []byte) Challenge {
	var c Challenge
	json.Unmarshal(b, &c)
	fmt.Println(c)
	return c
}

func getPayload(b []byte) payload {
	var p payload
	json.Unmarshal(b, &p)
	fmt.Println(p)
	return p
}

func replyChallenge(c Challenge) []byte {
	var cr ChallengeResponse
	var output []byte
	cr.Challenge = c.Challenge
	fmt.Println("Challenge accepted !")
	output, _ = json.Marshal(cr)
	return output
}
func initEnv() {
	botToken = os.Getenv("BOT_TOKEN")
	authorID = os.Getenv("AUTHOR_ID")
}
func managePayload(p payload) {
	fmt.Println("message :> " + p.Event.Text)
	if p.Event.Username != "ben-bot" && p.Event.Type == "app_mention" {
		if p.Event.User == authorID {
			//	sendMessage(p.Event.Channel, "<@"+p.Event.User+"> Oh non pas toi ! Il va encore tout faire péter !")
			//	sendMessage(p.Event.Channel, "de toute façon....")
		}
		if strings.Contains(p.Event.Text, "design") || strings.Contains(p.Event.Text, "doc") {
			sendMessage(p.Event.Channel, "<@"+p.Event.User+"> c'est quoi encore ce design ?! ")
			return
		}
		sendMessage(p.Event.Channel, "<@"+p.Event.User+">"+getRandomResponse())
	}
}

func sendMessage(channel, message string) {
	url := "https://slack.com/api/chat.postMessage" + "?token=" + botToken + "&channel=" + channel + "&text=" + url.PathEscape(message)
	color.White("Send message on slack...")

	// Send req using http Client
	slackResponse := GET(url)
	color.Green("slack response: " + string(slackResponse))
}

// GET - General function to send get REST Request
func GET(url string) []byte {
	color.Yellow("GET")
	color.Magenta("url: %s", url)
	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Send req using http Client
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("Error on response.\n[ERRO] -", err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if body != nil {
		output := []byte(body)
		color.Blue(string(output))
		return output
	}
	return nil
}

func getJoke() Joke {
	var j Joke
	r := GET("https://icanhazdadjoke.com/slack")
	fmt.Println(string(r))
	err := json.Unmarshal(r, &j)

	if err != nil {
		panic(err)
	}
	return j
}

func getRandomResponse() string {
	rand.Seed(time.Now().Unix())
	a := []string{
		"Il me faut un doc de design de Bot avant de pouvoir répondre !",
		"CALMEZ-VOUS !",
		"il parait qu'il y a une réponse sur confluence",
		"ouais mais en fait non !",
		"va voir ça avec <@" + authorID + ">...",
		"ma réponse sera 42",
		"quoi ? Moi ? non, j'ai rien à voir avec ça...",
		"ok",
		"Et bah tu sais quoi ? parquoi pas tiens !",
		"mmmh... On va demander au PO d'abord",
		"Quand il y aura à nouveau du cidre chez oscar",
		"tu vois comment épeler le mot lapin? Bah voilà c'est presque pareil",
		"attend je vais demander à <@UH6CBAXQD>...",
		"ou sinon, on s'écoute un petit Bob Marley OKLM",
		"Pas bête !",
		"Et tu as pensé ça tout seul ? bravo !",
		"Je m'énerve pas, J'EXPLIQUE !",
	}

	n := rand.Int() % len(a)

	fmt.Print("Gonna work from home...", a[n])
	return a[n]
}
