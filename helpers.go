package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-redis/redis/v8"
)

func getFresh() (fresh []*post) {
	IDs, err := rdb.ZRevRange(rdbctx, "STREAM:CHRON", 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	for _, ID := range IDs {
		t := &post{}

		_ = rdb.HGetAll(rdbctx, ID).Scan(t)
		if t.Publish == "TRUE" {
			fresh = append(fresh, t)
		}
	}
	return
}

func getHot() (hot []*post) {
	IDs, err := rdb.ZRevRangeByScore(rdbctx, "STREAM:HOT", &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  -1,
	}).Result()
	if err != nil {
		fmt.Println(err)
	}
	for _, ID := range IDs {
		t := &post{}

		_ = rdb.HGetAll(rdbctx, ID).Scan(t)
		if t.Publish == "TRUE" {
			hot = append(hot, t)
		}
	}
	return
}

func marshalPageData(r *http.Request) (*pageData, error) {
	t := &pageData{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(t)
	if err != nil {
		return t, err
	}
	return t, nil
}

// exeTmpl is used to build and execute an html template.
func exeTmpl(w http.ResponseWriter, r *http.Request, page *pageData, tmpl string) {
	// Add the user data to the page if they're logged in.
	c := r.Context().Value(ctxkey)
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		page.UserData = a

		err := templates.ExecuteTemplate(w, tmpl, page)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	err := templates.ExecuteTemplate(w, tmpl, page)
	if err != nil {
		fmt.Println(err)
	}
}

func getLikes(r *http.Request, name string) (likedTracks []*post) {
	// c := r.Context().Value(ctxkey)
	// if a, ok := c.(*credentials); ok && a.IsLoggedIn {
	likes, err := rdb.ZRange(rdbctx, name+":LIKES", 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}

	for _, trackID := range likes {
		t := &post{}

		_ = rdb.HGetAll(rdbctx, trackID).Scan(t)
		likedTracks = append(likedTracks, t)
	}

	// }
	return
}

func setLikes(r *http.Request, stream []*post) []*post {
	c := r.Context().Value(ctxkey)
	var user string
	if a, ok := c.(*credentials); ok && a.IsLoggedIn {
		user = a.Name
	} else {
		user = strings.Split(r.RemoteAddr, ":")[0]
	}
	for _, post := range stream {
		_, err := rdb.ZScore(rdbctx, user+":LIKES", post.ID).Result()
		if err != nil {
			post.Liked = false
		} else {
			post.Liked = true
		}
	}
	// return stream
	// for _, post := range stream {
	// 	post.Liked = false
	// }
	return stream
}

// marshalCredentials is used convert a request body into a credentials{}
// struct
func marshalCredentials(r *http.Request) (*credentials, error) {
	t := &credentials{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(t)
	if err != nil {
		return t, err
	}
	return t, nil
}

func marshalPostData(r *http.Request) (*post, error) {
	p := &post{}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(p)
	if err != nil {
		return p, err
	}
	return p, nil
}

// ajaxResponse is used to respond to ajax requests with arbitrary data in the
// format of map[string]string
func ajaxResponse(w http.ResponseWriter, res map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
	}
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

// makeZmem returns a redis Z member for use in a ZSET. Score is set to zero
func makeZmem(st string) *redis.Z {
	return &redis.Z{
		Member: st,
		Score:  0,
	}
}
