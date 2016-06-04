package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"text/template"
	"time"

	"github.com/gorilla/sessions"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

// BaseURL
const (
	BaseURL = "https://api.instagram.com"
)

// Instagram Client Id, Client Secret
var (
	ClientID     = os.Getenv("ClientID")
	ClientSecret = os.Getenv("ClientSecret")
	AccessToken  = os.Getenv("AccessToken")
	CallbackURL  = "https://stagram.guncy.com/oauth/callback"
)

// AuthorizeURL
var (
	AuthorizeURL   = BaseURL + "/oauth/authorize/?client_id=" + ClientID + "&scope=likes+comments&redirect_uri=" + CallbackURL + "&response_type=code"
	AccessTokenURL = BaseURL + "/oauth/access_token"
	PopularURL     = BaseURL + "/v1/media/popular?client_id=" + ClientID
	FeedURL        = BaseURL + "/v1/users/self/feed"
	LikedURL       = BaseURL + "/v1/users/self/media/liked"
	UserSearchURL  = BaseURL + "/v1/users/search"
	MediaURL       = BaseURL + "/v1/media"
	UserURL        = BaseURL + "/v1/users"
)

var store = sessions.NewCookieStore([]byte("NwrIYjjWXzQLoXtAgDhtmRHwp0I8WqvN"))

// SessionName ...
var SessionName = "stagram-session"

// Response ...
type Response struct {
	Pagination struct {
		NextURL       string `json:"next_url"`
		NextMaxLikeID string `json:"next_max_like_id"`
	} `json:"pagination"`
	Meta struct {
		code int `json:"code"`
	} `json:"meta"`
	Data []struct {
		ID   string `json:"id"`
		User struct {
			Username       string `json:"username"`
			FullName       string `json:"full_name"`
			ProfilePicture string `json:"profile_picture"`
			ID             string `json:"id"`
		} `json:"user"`
		Images struct {
			StandardResolution struct {
				URL string `json:"url"`
			} `json:"standard_resolution"`
		} `json:"images"`
		Likes struct {
			Count int `json:"count"`
		} `json:"likes"`
		Comments struct {
			Count int `json:"count"`
		} `json:"comments"`
	} `json:"data"`
}

// Users ...
type Users struct {
	Meta struct {
		code int `json:"code"`
	} `json:"meta"`
	Data []struct {
		Username       string `json:"username"`
		FullName       string `json:"full_name"`
		ProfilePicture string `json:"profile_picture"`
		ID             string `json:"id"`
	} `json:"data"`
}

// User ...
type User struct {
	Data struct {
		Username       string `json:"username"`
		FullName       string `json:"full_name"`
		ProfilePicture string `json:"profile_picture"`
		ID             string `json:"id"`
		Bio            string `json:"bio"`
		Website        string `json:"website"`
		Counts         struct {
			Media      int `json:"media"`
			Follows    int `json:"follows"`
			FollowedBy int `json:"followed_by"`
		} `json:"counts"`
	} `json:"data"`
}

// Media ...
type Media struct {
	Data struct {
		User struct {
			Username       string `json:"username"`
			ProfilePicture string `json:"profile_picture"`
			FullName       string `json:"full_name"`
			ID             string `json:"id"`
		} `json:"user"`
		Images struct {
			StandardResolution struct {
				URL string `json:"url"`
			} `json:"standard_resolution"`
		} `json:"images"`
		Likes struct {
			Count int `json:"count"`
			Data  []struct {
				Username       string `json:"username"`
				FullName       string `json:"full_name"`
				ProfilePicture string `json:"profile_picture"`
				ID             string `json:"id"`
			} `json:"data"`
		} `json:"likes"`
		Comments struct {
			Count int `json:"count"`
			Data  []struct {
				Text string `json:"text"`
				From struct {
					Username       string `json:"username"`
					FullName       string `json:"full_name"`
					ProfilePicture string `json:"profile_picture"`
					ID             string `json:"id"`
				} `json:"from"`
			} `json:"data"`
		} `json:"comments"`
		// Tags    []string `json:"tags"`
		Caption struct {
			Text string `json:"text"`
			From struct {
				Username       string `json:"username"`
				FullName       string `json:"full_name"`
				ProfilePicture string `json:"profile_picture"`
				ID             string `json:"id"`
			} `json:"from"`
		} `json:"caption"`
	} `json:"data"`
}

func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	goji.Get("/oauth/authorize", authorizeHandler)
	goji.Get("/oauth/callback", callbackHandler)
	goji.Get("/", homeHandler)
	goji.Get("/about", aboutHandler)
	goji.Get("/logout", logoutHandler)
	goji.Get("/feed", feedHandler)
	goji.Get("/like", likeHandler)
	goji.Get("/me", meHandler)
	goji.Get("/media/:media_id", mediaHandler)
	goji.Get("/ajax/more", moreHandler)
	goji.Get("/ajax/popular", morePopularHandler)
	goji.Get("/:username", userHandler)

	goji.Serve()
}

func authorizeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, AuthorizeURL, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	values := url.Values{}
	values.Add("client_id", ClientID)
	values.Add("client_secret", ClientSecret)
	values.Add("grant_type", "authorization_code")
	values.Add("redirect_uri", CallbackURL)
	values.Add("code", code)

	req, err := http.NewRequest("POST", AccessTokenURL, bytes.NewBufferString(values.Encode()))
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	setAccessToken(w, r, result["access_token"].(string))
	setUsername(w, r, result["user"].(map[string]interface{})["username"].(string))

	http.Redirect(w, r, "/feed", http.StatusFound)

}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	req, err := http.NewRequest("GET", PopularURL, nil)
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	response := &Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	result := map[string]interface{}{
		"Result": map[string]interface{}{
			"Response":   response,
			"IsLoggedIn": isLoggedIn(r),
			"Title":      "Popular",
			"Type":       "popular",
		},
	}

	tmpl, err := template.ParseFiles("list.html")
	if err != nil {
		panic(err)
	}

	if err = tmpl.Execute(w, result); err != nil {
		panic(err)
	}

}

func morePopularHandler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", PopularURL, nil)
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := getAccessToken(r)

	req, err := http.NewRequest("GET", LikedURL+"?count=33&access_token="+accessToken, nil)
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	response := &Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	result := map[string]interface{}{
		"Result": map[string]interface{}{
			"Response":   response,
			"IsLoggedIn": isLoggedIn(r),
			"Title":      "Like",
			"Type":       "like",
		},
	}

	tmpl, err := template.ParseFiles("list.html")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, result)
	if err != nil {
		panic(err)
	}

}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := getAccessToken(r)
	req, err := http.NewRequest("GET", FeedURL+"?count=33&access_token="+accessToken, nil)
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	response := &Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	result := map[string]interface{}{
		"Result": map[string]interface{}{
			"Response":   response,
			"IsLoggedIn": isLoggedIn(r),
			"Title":      "Feed",
			"Type":       "feed",
		},
	}

	tmpl, err := template.ParseFiles("list.html")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(w, result)
	if err != nil {
		panic(err)
	}

}

func moreHandler(w http.ResponseWriter, r *http.Request) {

	nextURL := r.URL.Query().Get("next_url")
	count := r.URL.Query().Get("count")
	maxLikeID := r.URL.Query().Get("max_like_id")
	maxID := r.URL.Query().Get("max_id")

	req, err := http.NewRequest("GET", nextURL+"&count="+count+"&max_like_id="+maxLikeID+"&max_id="+maxID, nil)
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

}

func userHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	username := c.URLParams["username"]

	user, err := getUser(r, username)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	userID := user.Data.ID

	accessToken := getAccessToken(r)
	if accessToken == "" {
		accessToken = AccessToken
	}

	// Media
	req, err := http.NewRequest("GET", BaseURL+"/v1/users/"+userID+"/media/recent?&access_token="+accessToken, nil)
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	response := &Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	result := map[string]interface{}{
		"Result": map[string]interface{}{
			"Response":   response,
			"User":       user,
			"IsLoggedIn": isLoggedIn(r),
			"Title":      username,
			"Type":       "user",
		},
	}

	tmpl, err := template.ParseFiles("list.html")
	if err != nil {
		panic(err)
	}

	if err = tmpl.Execute(w, result); err != nil {
		panic(err)
	}

}

func meHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	http.Redirect(w, r, "/"+username, http.StatusFound)
}

func mediaHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	mediaID := c.URLParams["media_id"]

	accessToken := getAccessToken(r)
	if accessToken == "" {
		accessToken = AccessToken
	}

	// Media
	req, err := http.NewRequest("GET", MediaURL+"/"+mediaID+"?access_token="+accessToken, nil)
	if err != nil {
		panic(err)
	}

	body := doRequest(req)

	response := &Media{}
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println(err)
		panic(err)
	}

	result := map[string]interface{}{
		"Result": map[string]interface{}{
			"Response":   response,
			"IsLoggedIn": isLoggedIn(r),
			"Title":      response.Data.User.Username,
			"Type":       "media",
		},
	}

	tmpl, err := template.ParseFiles("media.html")
	if err != nil {
		panic(err)
	}

	if err = tmpl.Execute(w, result); err != nil {
		panic(err)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	setAccessToken(w, r, "")
	http.Redirect(w, r, "/", http.StatusFound)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("about.html")
	if err != nil {
		panic(err)
	}

	result := map[string]interface{}{
		"Result": map[string]interface{}{
			"IsLoggedIn": isLoggedIn(r),
			"Title":      "About",
		},
	}

	if err = tmpl.Execute(w, result); err != nil {
		panic(err)
	}

}

// Private Funcs
func getAccessToken(r *http.Request) string {
	session, _ := store.Get(r, SessionName)
	if session.Values["access_token"] == nil {
		return ""
	}
	return session.Values["access_token"].(string)
}

func setAccessToken(w http.ResponseWriter, r *http.Request, value string) {
	session, _ := store.Get(r, SessionName)
	session.Values["access_token"] = value
	session.Save(r, w)
}

func isLoggedIn(r *http.Request) bool {
	if getAccessToken(r) == "" {
		return false
	}
	return true
}

func getUsername(r *http.Request) string {
	session, _ := store.Get(r, SessionName)
	if session.Values["user_name"] == nil {
		return ""
	}
	return session.Values["user_name"].(string)
}

func setUsername(w http.ResponseWriter, r *http.Request, username string) {
	session, _ := store.Get(r, SessionName)
	session.Values["user_name"] = username
	session.Save(r, w)
}

func getUser(r *http.Request, username string) (*User, error) {
	accessToken := getAccessToken(r)
	if accessToken == "" {
		accessToken = AccessToken
	}

	searchReq, err := http.NewRequest("GET", UserSearchURL+"?q="+username+"&count=1&access_token="+accessToken, nil)
	if err != nil {
		return nil, err
	}

	body := doRequest(searchReq)

	users := &Users{}
	if err := json.Unmarshal(body, users); err != nil {
		return nil, err
	}
	userID := users.Data[0].ID
	getReq, err := http.NewRequest("GET", UserURL+"/"+userID+"/?access_token="+accessToken, nil)
	if err != nil {
		return nil, err
	}

	body = doRequest(getReq)

	user := &User{}
	if err := json.Unmarshal(body, user); err != nil {
		return nil, err
	}

	return user, nil
}

func doRequest(req *http.Request) []byte {
	client := &http.Client{Timeout: time.Duration(30) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return body

}
