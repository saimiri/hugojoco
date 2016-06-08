/*
Hugocs or Hugo Comment Script is a script that processes POST requests
in a specific format and converts them into JSON files that are saved
to the filesystem. These files can then be processed using Hugo's
readDir and getJSON functions and displayed as comments.

Copyright 2016 Juha Auvinen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Comment struct {
	Name           string `json:"name"`
	Email          string `json:"-"`
	EmailMd5       string `json:"emailMd5"`
	EmailMd5Salted string `json:"emailMd5Salted"`
	Website        string `json:"website"`
	AvatarType     string `json:"avatarType"`
	IPAddress      string `json:"ipv4Address"`
	PageID         string `json:"pageId"`
	Body           string `json:"body"`
	Timestamp      string `json:"timestamp"`
}

type Config struct {
	BaseDir     string
	CommentsDir string
	ContentDir  string
	Salt        string
	TouchFile   string
}

type Response struct {
	Message string `json:"message"`
	IsError bool   `json:"isError"`
}

var config = Config{
	BaseDir:     ".",
	CommentsDir: "comments",
	ContentDir:  "content",
	TouchFile:   ".comment"}

// Used for testing
func newComment(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("new.html")
	t.Execute(w, "")
}

func saveComment(w http.ResponseWriter, r *http.Request) {
	// TODO: Allow other content-types for javascript-disabled clients.
	// Maybe render a HTML template?
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Message: "",
		IsError: true}

	if r.Method == "POST" {
		r.ParseForm()
		_, err := validateComment(r, config.ContentDir)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = err.Error()
		} else {
			now := time.Now()
			comment := Comment{
				Name:           r.Form.Get("name"),
				Email:          r.Form.Get("email"),
				EmailMd5:       getMd5(r.Form.Get("email"), ""),
				EmailMd5Salted: getMd5(r.Form.Get("email"), config.Salt),
				Website:        r.Form.Get("website"),
				AvatarType:     r.Form.Get("avatar_type"),
				IPAddress:      getIPAddress(r),
				PageID:         r.Form.Get("page_id"),
				Body:           processBody(r.Form.Get("body")),
				Timestamp:      now.Format(time.RFC3339)}

			// FIXME: Needs error handling
			jsonData, _ := json.Marshal(comment)
			filename := buildFilename(comment.Name, comment.Body, buildTimestamp(now))
			writePath := filepath.Join(config.CommentsDir, comment.PageID)
			writeCommentToDisk(jsonData, writePath, filename)

			//str := fmt.Sprintf( "%#v", r )

			w.WriteHeader(http.StatusOK)
			response.Message = "Thank you for the comment"
			response.IsError = false
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		response.Message = "Must be POST"
	}
	// FIXME: Needs error handling
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// TODO: Check field lengths to prevent too large comment files
// TODO: Error messages should be read from a config file
func validateComment(r *http.Request, contentDir string) (string, error) {
	form := r.Form
	if len(form.Get("last_name")) > 0 {
		return "last_name", errors.New("You appear to be a spammer, or your browser auto-fills this form.")
	}
	if len(form.Get("name")) > 128 || regexp.MustCompile(`(?i)[\w\s\d\-]+`).MatchString(form.Get("name")) != true {
		return "name", errors.New("Name is not valid")
	}
	if len(form.Get("email")) > 128 || regexp.MustCompile(`^[a-z0-9_.\-\+]+@[a-z0-9_.\-\+]+$`).MatchString(form.Get("email")) != true {
		return "email", errors.New("Email address is not valid")
	}
	if len(form.Get("website")) > 128 || form.Get("website") != "" && regexp.MustCompile(`(https?:\/\/)?[a-z0-9\-\.]+`).MatchString(form.Get("website")) != true {
		return "website", errors.New("Website is not valid")
	}
	if len(form.Get("avatar_type")) > 32 || regexp.MustCompile(`[a-z]+`).MatchString(form.Get("avatar_type")) != true {
		return "avatar_type", errors.New("Avatar type is not valid")
	}
	if len(form.Get("page_id")) > 1024 || regexp.MustCompile(`[a-z0-9\-]+(\/[a-z0-9\-]+)*`).MatchString(form.Get("page_id")) != true {
		return "page_id", errors.New("page_id is not valid")
	}
	if len(form.Get("content_type")) > 4 || regexp.MustCompile(`[a-z]+`).MatchString(form.Get("content_type")) != true {
		return "content_type", errors.New("Content type is not valid")
	}
	if len(form.Get("body")) > 8192 || form.Get("body") == "" {
		return "body", errors.New("You forgot to write the actual comment!")
	}
	if !postExists(form.Get("page_id"), form.Get("content_type"), contentDir) {
		return "page_id", errors.New("Specified post does not exist")
	}
	return "", nil
}

func isValidPath(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	//panic(path)
	return false
}

func postExists(pageID string, extension string, contentDir string) bool {
	return isValidPath(filepath.Join(contentDir, pageID+"."+extension))
}

func processBody(body string) string {
	body = strings.Replace(body, `"""`, "%quote%", -1)
	body = template.HTMLEscapeString(body)
	body = strings.Replace(body, "%quote%", ">", -1)
	return body
}

// getMd5 generates a md5 checksum for given string with optional salt.
// Used here to hash the email address for avatar service URLs.
func getMd5(s string, salt string) string {
	hash := md5.Sum([]byte(s + salt))
	return hex.EncodeToString(hash[:])
}

func getIPAddress(r *http.Request) string {
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	// FIXME: This should probably look at other fields instead
	return "Unknown"
}

func debugJSON(comment Comment) string {
	str := fmt.Sprintf("%#v", comment)
	return str
}

func buildFilename(name string, msg string, timestamp string) string {
	firstWords := getFirstWords(name + " " + msg)
	return timestamp + "-" + firstWords + ".json"
}

func getFirstWords(s string) string {
	re, _ := regexp.Compile(`(?i)[a-z]+`)
	res := re.FindAllStringSubmatch(s, 7)
	parts := make([]string, 0)
	for _, part := range res {
		parts = append(parts, strings.ToLower(part[0]))
	}
	joined := strings.Join(parts, "-")
	if len(joined) > 32 {
		joined = joined[:32]
	}
	return strings.TrimRight(joined, " -")
}

/*
BuildTimestamp builds a timestamp string to be used as part of the
filename for JSON comment files.
*/
func buildTimestamp(now time.Time) string {
	dateString := fmt.Sprintf("%d-%02d-%02d-%02d%02d%02d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	return dateString
}

func writeCommentToDisk(JSON []byte, path string, filename string) error {
	// If path doesn't exist, create it
	if !isValidPath(path) {
		os.MkdirAll(path, os.ModePerm)
	}
	fn := filepath.Join(path, filename)
	// Write file to disk
	// FIXME: Needs error handling
	ioutil.WriteFile(fn, JSON, 0600)
	if config.TouchFile != "" {
		updateChangeFile(filepath.Join(config.TouchFile))
	}
	return nil
}

func updateChangeFile(touchFile string) {
	// FIXME: Needs error handling
	ioutil.WriteFile(touchFile, []byte("."), 0600)
}

func main() {
	sourceFlag := flag.String("src", ".", "Full path to the sources dir.")
	contentFlag := flag.String("content", "content", "The content directory. Relative to source dir.")
	commentsFlag := flag.String("comments", "comments", "The directory to save comments. Relative to source dir.")
	touchFlag := flag.String("touch", "", "File to update when a new comment directory is created. Relative to source dir. Some watch scripts may require this.")
	addressFlag := flag.String("address", "", "IP address to use for incoming requests. Defaults to any address.")
	portFlag := flag.String("port", "8080", "Port to listen to for incoming requests.")
	pathFlag := flag.String("path", "/comment", "The url path that is used for comment processing.")
	saltFlag := flag.String("salt", "", "Salt used when generating hashes for anonymous email addresses")

	flag.Parse()

	config.BaseDir = *sourceFlag
	config.ContentDir = filepath.Join(*sourceFlag, *contentFlag)
	config.CommentsDir = filepath.Join(*sourceFlag, *commentsFlag)
	if *touchFlag != "" {
		config.TouchFile = filepath.Join(*sourceFlag, *touchFlag)
	} else {
		config.TouchFile = ""
	}
	config.Salt = *saltFlag

	serverAddress := *addressFlag + ":" + *portFlag

	http.HandleFunc(*pathFlag, saveComment)
	http.HandleFunc("/new", newComment)
	http.ListenAndServe(serverAddress, nil)
}
