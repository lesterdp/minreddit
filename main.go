package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var t string
var tpl = template.Must(template.ParseFiles("index.html"))

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type Search struct {
	Searchkey string
	NumPost   int
	Results   Subreddit
}

type Subreddit struct {
	Data *Subredditdata `json:"data"`
}

type Subredditdata struct {
	Modhash  string  `json:"modhash"`
	Dist     int     `json:"dist"`
	Children []Posts `json:"children"`
}

type Posts struct {
	Kind string    `json:"kind"`
	Data *Postdata `json:"data"`
}

type Postdata struct {
	Subreddit string `json:"subreddit"`
	Selftext  string `json:"selftext"`
	Author    string `json:"author"`
	Title     string `json:"title"`
	Url       string `json:"url"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	search := &Search{}

	search.NumPost = 20
	requestURL := fmt.Sprintf("https://oauth.reddit.com/r/popular?g=us&limit=%d&raw_json=1", search.NumPost)
	var bearer = "Bearer " + t

	request, err := http.NewRequest("GET", requestURL, nil)

	request.Header.Add("Authorization", bearer)
	request.Header.Add("User-Agent", "Deric")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error on response.\n[ERRO] -", err)
	}

	err = json.NewDecoder(response.Body).Decode(&search.Results)
	if err != nil {
		log.Fatalln(err)
	}

	err = tpl.Execute(w, search)

}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")
	page := params.Get("page")
	if page == "" {
		page = "1"
	}

	search := &Search{}
	search.Searchkey = searchKey

	search.NumPost = 20
	requestURL := fmt.Sprintf("https://oauth.reddit.com/r/%s/hot?g=us&limit=%d&raw_json=1", url.QueryEscape(search.Searchkey), search.NumPost)
	var bearer = "Bearer " + t

	request, err := http.NewRequest("GET", requestURL, nil)

	request.Header.Add("Authorization", bearer)
	request.Header.Add("User-Agent", "Deric")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error on response.\n[ERRO] -", err)
	}

	err = json.NewDecoder(response.Body).Decode(&search.Results)
	if err != nil {
		log.Fatalln(err)
	}

	err = tpl.Execute(w, search)

}

func main() {
	t = (CreateAccessToken("clientID", "clientSecret"))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)
}

func CreateAccessToken(clientID string, clientSecret string) string {
	requestURL := "https://www.reddit.com/api/v1/access_token"
	body := strings.NewReader("grant_type=client_credentials")

	request, err := http.NewRequest(http.MethodPost, requestURL, body)
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("User-Agent", "Deric")
	request.SetBasicAuth(clientID, clientSecret)
	//request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := new(http.Client)
	response, err := httpClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	tokenData := Token{}
	err = json.NewDecoder(response.Body).Decode(&tokenData)
	if err != nil {
		log.Fatal(err)
	}
	//println(tokenData.AccessToken)
	return tokenData.AccessToken
}
