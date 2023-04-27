package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func latestView(w http.ResponseWriter, r *http.Request) {
	var page pageData
	stream := getFresh()
	page.Stream = setLikes(r, stream)
	page.Company = websiteName
	page.Stream = stream
	page.UserData = &credentials{}
	page.PageName = "LATEST POSTS"
	exeTmpl(w, r, &page, "main.tmpl")
}

func hotView(w http.ResponseWriter, r *http.Request) {
	var page pageData
	ts := getHot()
	page.Company = websiteName
	page.Stream = setLikes(r, ts)
	page.UserData = &credentials{}
	page.PageName = "HOTTEST POSTS"
	exeTmpl(w, r, &page, "main.tmpl")
}

func trackView(w http.ResponseWriter, r *http.Request) {
	trackID := strings.Split(r.URL.Path, "/")[2]
	var p []*post
	var page pageData
	t := &post{}
	rdb.HGetAll(rdbctx, trackID).Scan(t)
	p = append(p, t)
	page.Company = websiteName
	page.Stream = setLikes(r, p)
	page.UserData = &credentials{}
	page.PageName = ""
	exeTmpl(w, r, &page, "main.tmpl")
}

func likesView(w http.ResponseWriter, r *http.Request) {
	name := strings.Split(r.URL.Path, "/")[2]
	var page pageData
	page.Company = websiteName
	page.Stream = setLikes(r, getLikes(r, name))
	page.UserData = &credentials{}
	page.PageName = name + "'s Liked Posts"
	exeTmpl(w, r, &page, "main.tmpl")
}

func getStream(w http.ResponseWriter, r *http.Request) {
	page, err := marshalPageData(r)
	if err != nil {
		log.Println(err)
	}
	page.Company = websiteName

	var stream []*post
	if page.Category == "LATEST" {
		stream = getFresh()
		page.PageName = "LATEST POSTS"
	} else if page.Category == "HOT" {
		stream = getHot()
		page.PageName = "HOTTEST POSTS"
	} else if page.Category == "STREAM" {
		// Get users followed tags stream
		// See what other users are liking in real time
		// TODO
	} else if page.Category == "TRACK" {
		fmt.Println(page.Number, page)
		t := &post{}
		err := rdb.HGetAll(rdbctx, page.Number).Scan(t)
		if err != nil {
			log.Println(err)
		}

		stream = append(stream, t)
		page.PageName = ""
	} else {
		stream = getLikes(r, page.Category)
		page.PageName = page.Category + "'s Liked Posts"
	}
	page.Stream = setLikes(r, stream)

	c := r.Context().Value(ctxkey)
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		page.UserData = c.(*credentials)
	} else {
		page.UserData = &credentials{}
	}

	var b bytes.Buffer
	err = templates.ExecuteTemplate(&b, "updatePage.tmpl", page)
	if err != nil {
		fmt.Println(err)
	}
	j, err := json.Marshal(page.Stream)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
	ajaxResponse(w, map[string]string{
		"success":   "true",
		"error":     "false",
		"tracklist": string(j),
		"template":  b.String(),
	})
}

func likePost(w http.ResponseWriter, r *http.Request) {
	pd, err := marshalPostData(r)
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "Error parsing data",
		})
		return
	}
	// Check if the user is logged in. In this app we also allow anonymous
	// posting with an IP address as the user.
	c := r.Context().Value(ctxkey)
	var user string
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		user = a.Name
	} else {
		user = strings.Split(r.RemoteAddr, ":")[0]
	}

	pipe := rdb.Pipeline()
	zmem := makeZmem(pd.ID)
	result, err := rdb.ZAdd(rdbctx, user+":LIKES", zmem).Result()
	if err != nil {
		fmt.Println(err)
	}

	// If the track is already in the users LIKES, we remove it,
	// and decrement the score
	if result == 0 {
		_, err := rdb.ZRem(rdbctx, user+":LIKES", pd.ID).Result()
		if err != nil {
			log.Print(err)
		}

		_, err = rdb.ZIncrBy(rdbctx, "STREAM:HOT", -1, pd.ID).Result()
		if err != nil {
			log.Print(err)
		}

		ajaxResponse(w, map[string]string{
			"success": "true",
			"isLiked": "false",
			"error":   "false",
		})
		return
	}

	// pipe.ZIncr(rdbctx, "STREAM:ALL", zmem).Result()
	pipe.ZIncrBy(rdbctx, "STREAM:HOT", 1, pd.ID).Result()
	_, err = pipe.Exec(rdbctx)
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"isLiked": "",
			"error":   "Error updating database",
		})
		return

	}

	ajaxResponse(w, map[string]string{
		"success": "true",
		"isLiked": "true",
		"error":   "false",
	})
}

func newPost(w http.ResponseWriter, r *http.Request) {
	pd, err := marshalPostData(r)
	if err != nil {
		fmt.Println(err)
		ajaxResponse(w, map[string]string{
			"success": "false",
			"error":   "Error parsing data",
		})
		return
	}
	// Check if the user is logged in. You can't make a post without being
	// logged in.
	c := r.Context().Value(ctxkey)
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		id := genPostID(9)
		pipe := rdb.Pipeline()
		pipe.HMSet(rdbctx, id, pd)

		_, err = pipe.Exec(rdbctx)
		if err != nil {
			fmt.Println(err)
			ajaxResponse(w, map[string]string{
				"success": "false",
				"error":   "Error updating database",
			})
			return

		}

		ajaxResponse(w, map[string]string{
			"success": "true",
			"error":   "false",
		})
	}
}
