package main

import (
	"context"
	"html/template"
	"log"
	"math/rand"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/hartsfield/ohsheet"
)

var (
	// connect to redis
	redisIP = os.Getenv("redisIP")
	rdb     = redis.NewClient(&redis.Options{
		Addr:     redisIP + ":6379",
		Password: "",
		DB:       0,
	})

	rdx = context.Background()
)

// post is a post in a stream of posts
type post struct {
	Title    string        `redis:"Title" json:"title"`
	Body     template.HTML `redis:"Body" json:"body"`
	Image    string        `redis:"Image" json:"image"`
	Audio    string        `redis:"Audio" json:"audioMedia"`
	Video    string        `redis:"Video" json:"videoMedia"`
	ID       string        `redis:"ID" json:"ID"`
	Author   string        `redis:"Author" json:"author"`
	Parent   string        `redis:"Parent" json:"parent"`
	TS       string        `redis:"TS" json:"timestamp"`
	Tags     []string      `redis:"Tags" json:"tags"`
	Testing  string        `redis:"Testing" json:"testing"`
	Children []*post       `redis:"Children" json:"children"`
	Likes    int           `redis:"Likes" json:"likes"`
	Liked    bool          `redis:"Liked" json:"liked"`
}

func main() {
	pipe := rdb.Pipeline()
	for _, row := range getDataFromSheets() {
		post, ID := makePost(row)

		_, err := pipe.HMSet(rdx, ID, post).Result()
		if err != nil {
			log.Println(err)
		}

		_, err = pipe.ZAdd(rdx, "STREAM:CHRON", makeZmem(ID)).Result()
		if err != nil {
			log.Println(err)
		}
		_, err = pipe.ZAdd(rdx, "STREAM:HOT", makeZmem(ID)).Result()
		if err != nil {
			log.Println(err)
		}

	}
	_, err := pipe.Exec(rdx)
	if err != nil {
		log.Println(err)
	}

}

// makeZmem returns a redis Z member for use in a ZSET. Score is set to zero
func makeZmem(st string) *redis.Z {
	return &redis.Z{
		Member: st,
		Score:  0,
	}
}

func makePost(row []interface{}) (post map[string]interface{}, ID string) {
	ID = genPostID(9)
	return map[string]interface{}{
		"Title": row[0].(string),
		"Body":  row[1].(string),
		"Image": "../../public/assets/images/post-imgs/" + row[2].(string) + ".jpg",
		"ID":    ID,
	}, ID
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

func getDataFromSheets() [][]interface{} {
	// Connect to the API
	sheet := &ohsheet.Access{
		Token:       "token.json",
		Credentials: "credentials.json",
		Scopes:      []string{"https://www.googleapis.com/auth/spreadsheets"},
	}
	srv := sheet.Connect()

	spreadsheetId := "1oSv_qkovpTgkDnLAf-ucSYpgybqn6OkVDHgTzpJBvvw"
	readRange := "A2:C"

	resp, err := sheet.Read(srv, spreadsheetId, readRange)
	if err != nil {
		log.Panicln("Unable to retrieve data from sheet: ", err)
	}

	return resp.Values
}
