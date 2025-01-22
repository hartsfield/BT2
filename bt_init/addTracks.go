package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// track is an audio track
type track struct {
	Artist       string   `json:"artist"`
	Title        string   `json:"title"`
	Album        string   `json:"album"`
	Year         string   `json:"year"`
	Permission   string   `json:"permission"`
	Featuring    string   `json:"featuring"`
	Remix        string   `json:"remix"`
	Instrumental string   `json:"instrumental"`
	Origin       string   `json:"origin"`
	Publish      string   `json:"publish"`
	Image        string   `json:"image"`
	Path         string   `json:"path"`
	ID           string   `json:"id"`
	Likes        int      `json:"likes"`
	Liked        bool     `json:"liked"`
	LikedBy      []string `redis:"LikedBy" json:"likedBy"`
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		log.Println(err)
	}
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	// spreadsheetId := "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
	spreadsheetId := "1xEuxgdRLZTTIDoH54lLNKWHv-QXVbQcnDYVly6bTWxY"
	readRange := "Sheet1!A2:J"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for _, row := range resp.Values {
			if len(row) > 1 {
				formatted := strings.Join(strings.Split(row[0].(string)+" "+row[1].(string), " "), "_")
				if !checkExists(formatted) {
					var instrumental string
					if row[7].(string) == "true" {
						instrumental = "instrumental"
					}
					if row[2] == "single" {
						row[2] = ""
					}
					nodes, _ := getMarkup(makeLink(
						strings.Join(
							[]string{
								row[0].(string),
								row[1].(string),
								row[2].(string),
								instrumental,
							}, "+",
						)))
					var link string
					var art string
					for _, node := range nodes {
						link = findLink(node)
					}
					if len(link) > 3 {
						nodes, _ = getMarkup(link)
						for _, node := range nodes {
							art = parseArtLink(node)
						}
					}
					getArt(art, row[0].(string), row[1].(string))
					downloadLink(link, row, srv)
				}
			}
		}
	}
	addToRedis(resp.Values)
}

func downloadLink(link string, row []interface{}, srv *sheets.Service) {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	dirname := "../public/assets/track_data"
	log.Println(path[len(path)-len(dirname)+2:])
	if path[len(path)-len(dirname)+2:] != dirname[2:] {
		err := os.Chdir(dirname)
		if err != nil {
			log.Println("Couldn't change directory: ", dirname)
			return
		}
	}

	// Use youtube-dl to download the songs.
	// TODO: add native support for song downloading.
	fmt.Printf("Using youtube-dl to download %s...\n", link)
	// The following line splits the artist and song name by the spaces,
	// then joins them with underscores:
	// The Beatles Let It Be >> The_Beatles_Let_It_Be
	formatted := strings.Join(strings.Split(row[0].(string)+" "+row[1].(string), " "), "_")
	c := exec.Command("youtube-dl", "--output="+formatted+".mp3", link)

	err = c.Run()
	if err != nil {
		log.Println("Something went wrong starting or running youtube-dl. Are you sure it's installed and that it's in your $PATH?: ", err)
		// spreadsheetId := "1xEuxgdRLZTTIDoH54lLNKWHv-QXVbQcnDYVly6bTWxY"
		// writeRange := "Sheet1!K2"
		// rb := &sheets.ValueRange{
		// 	// TODO: Add desired fields of the request body. All existing fields
		// 	// will be replaced.
		// }
		// resp, err := srv.Spreadsheets.Values.Update(spreadsheetId, writeRange, rb).Do()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// log.Println(resp)
	}
	log.Println("Downloaded", link)
}

func checkExists(file string) bool {
	if _, err := os.Stat("../public/assets/track_data/" + file + ".jpg"); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func getArt(artLink, artist, title string) {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	dirname := "../public/assets/track_data"
	if path[len(path)-len(dirname)+2:] != dirname[2:] {
		err := os.Chdir(dirname)
		if err != nil {
			log.Println("Couldn't change directory: ", dirname)
			return
		}
	}

	r, err := http.Get(artLink)
	if err != nil {
		fmt.Println("Couldn't download image from link: ", err)
		return
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body, this should not effect anything: ", err)
		}
	}()

	// Create the file, we call it albumArt and append the extension from the
	// downloaded file to it.
	file, err := os.Create(strings.Join(strings.Split(artist+" "+title, " "), "_") + ".jpg")
	if err != nil {
		fmt.Println("Error creating album art file: ", err)
		return
	}

	// Save the data to the file
	_, err = io.Copy(file, r.Body)
	if err != nil {
		fmt.Println("Error saving album art: ", err)
	}

	err = file.Close()
	if err != nil {
		log.Println("Couldn't close album art file: ", err)
	}
	// os.Chdir("..")
	log.Println("Downloaded art for", artist, title)
}

func parseArtLink(n *html.Node) string {
	var link string
	var parseFunc func(*html.Node)
	parseFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" && link == "" {
			for l, a := range n.Attr {
				if a.Key == "class" && a.Val == "popupImage" {
					link = n.Attr[l+1].Val
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseFunc(c)
		}
	}
	parseFunc(n)
	return link
}

var (
	// connect to redis
	redisIP = os.Getenv("redisIP")
	rdb     = redis.NewClient(&redis.Options{
		Addr:     redisIP + ":6379",
		Password: "",
		DB:       5,
	})

	// this context is used for the client/server connection. It's useful
	// for passing the token/credentials around.
	rdbctx = context.Background()
)

// makeZmem returns a redis Z member for use in a ZSET. Score is set to zero
func makeZmem(st string) *redis.Z {
	return &redis.Z{
		Member: st,
		Score:  0,
	}
}

// makeTrack takes a directory entry and uses the information to return a track
// object
func makeTrack(file fs.DirEntry) *track {
	fn := strings.Split(file.Name(), "-")[1]
	fn = fn[:strings.LastIndex(fn, ".")]
	return addImage(&track{
		Artist: strings.Split(file.Name(), "-")[0],
		Title:  fn,
		Path:   file.Name(),
		ID:     genPostID(5),
	})
}

// addImage adds an image to a track object by analyzing the first 6 characters
// of the image file name and audio file name and checking for a match. For this
// to work, audio and images must be added to the appropriate directory with
// intent
func addImage(t *track) *track {
	images, _ := os.ReadDir("../public/assets/images")
	for _, image := range images {
		if strings.ToLower(image.Name())[0:6] == strings.ToLower(t.Artist[0:6]) {
			t.Image = image.Name()
		}
	}
	return t
}

// genPostID generates a post ID
func genPostID(length int) (ID string) {
	symbols := "abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := 0; i <= length; i++ {
		s := rand.Intn(len(symbols))
		ID += symbols[s : s+1]
	}
	return
}

func addToRedis(vals [][]interface{}) {
	pipe := rdb.Pipeline()
	for _, row := range vals {
		ID := genPostID(9)
		if len(row) > 1 {
			formated := strings.Join(strings.Split(row[0].(string)+" "+row[1].(string), " "), "_")
			_, err := pipe.HMSet(rdbctx, ID, map[string]interface{}{
				"Artist":       row[0].(string),
				"Title":        row[1].(string),
				"Album":        row[2].(string),
				"Year":         row[3].(string),
				"Permission":   row[4].(string),
				"Featuring":    row[5].(string),
				"Remix":        row[6].(string),
				"Instrumental": row[7].(string),
				"Origin":       row[8].(string),
				"Publish":      fmt.Sprint(row[9]),
				"Path":         "../public/assets/track_data/" + formated + ".mp3",
				"Image":        "../public/assets/track_data/" + formated + ".jpg",
				"ID":           ID,
				"Likes":        "0",
				"Liked":        false,
			}).Result()
			if err != nil {
				fmt.Println(err)
			}

			zm := makeZmem(ID)

			pipe.ZAdd(rdbctx, "STREAM:CHRON", zm)
			pipe.ZAdd(rdbctx, "STREAM:HOT", zm)
			_, err = pipe.Exec(rdbctx)
			if err != nil {
				log.Println(err)
			}
		}
	}
	fmt.Println("test")
}

func makeLink(keywords string) string {
	link := "https://bandcamp.com/search?q="
	keywords = strings.Join(strings.Split(keywords, " "), "+")
	return link + keywords
}

// getMarkup is used to obtain the html text and actual node objects that will
// be parsed.
func getMarkup(link string) ([]*html.Node, []string) {
	var nodes []*html.Node
	var text []string
	l := strings.Replace(link, "https", "http", 1)
	r, err := http.Get(l)
	if err != nil {
		log.Panicln(err, `Looks like the page wouldn't load`)
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body, this should not effect anything: ", err)
		}
	}()

	buf, _ := io.ReadAll(r.Body)
	// use nopCloser to get an io.readCloser which implements the io.Reader
	// interface used for html.Parse().
	rw1 := io.NopCloser(bytes.NewBuffer(buf))
	n, err := html.Parse(rw1)
	if err != nil {
		fmt.Println(err)
	}
	nodes = append(nodes, n)
	text = append(text, string(buf))
	return nodes, text
}

func findLink(n *html.Node) string {
	var link string
	var parseFunc func(*html.Node)
	parseFunc = func(n *html.Node) {
		if n.Data == "a" && link == "" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					if strings.Contains(a.Val, "/track/") || strings.Contains(a.Val, "/album/") {
						link = strings.SplitAfter(a.Val, "?")[0]
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseFunc(c)
		}
	}
	parseFunc(n)
	return link
}
