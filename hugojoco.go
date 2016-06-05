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
	Name       string
	Email      string
	Website    string
	GravatarID string
	IPAddress  string
	PostID     string
	Body       string
	Timestamp  string
}

type Config struct {
	BaseDir string
	CommentsDir string
	ContentDir string
	TouchFile string
}

type Response struct {
	Message string
	IsError bool
}

var config = Config {
	BaseDir: ".",
	CommentsDir: "comments",
	ContentDir: "content",
	TouchFile: ".comment" }

// Used for testing
func newComment(w http.ResponseWriter, r *http.Request ) {
	t, _ := template.ParseFiles( "new.html" );
	t.Execute(w, "")
}

func saveComment(w http.ResponseWriter, r *http.Request ) {
	// TODO: Allow other content-types for javascript-disabled clients.
	// Maybe render a HTML template?
	w.Header().Set("Content-Type", "application/json")
	response := Response {
		Message: "",
		IsError: true }

	if r.Method == "POST" {
		r.ParseForm()
		_, err := validateComment(r, config.ContentDir)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response = Response{
				Message: err.Error(),
				IsError: true}
		} else {
			now := time.Now()
			comment := Comment{
				Name:       r.Form.Get("name"),
				Email:      r.Form.Get("email"),
				Website:    r.Form.Get("website"),
				GravatarID: getGravatarId(r.Form.Get("email")),
				IPAddress:  getIPAddress(r),
				PostID:     r.Form.Get("post_id"),
				Body:       processBody(r.Form.Get("body")),
				Timestamp:  now.Format(time.RFC3339)}

			// FIXME: Needs error handling
			jsonData, _ := json.Marshal(comment)
			filename := buildFilename(comment.Email, buildTimestamp(now))
			writePath := filepath.Join(config.CommentsDir, comment.PostID)
			writeCommentToDisk(jsonData, writePath, filename)

			//str := fmt.Sprintf( "%#v", r )
			
			w.WriteHeader(http.StatusOK)
			response = Response{
				Message: "Thank you for the comment",
				IsError: false}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		response = Response{
			Message: "Must be POST",
			IsError: true}
	}
	// FIXME: Needs error handling
	responseJSON, _ := json.Marshal(response)
	w.Write(responseJSON)
}

// TODO: Check field lengths to prevent too large comment files
// TODO: Error messages should be read from a config file
func validateComment(r *http.Request, contentDir string) (string, error) {
	form := r.Form
	if form.Get("last_name") != "" {
		return "last_name", errors.New("You appear to be a spammer, or your browser auto-fills this form.")
	}
	if regexp.MustCompile(`(?i)[\w\s\d]+`).MatchString(form.Get("name")) != true {
		return "name", errors.New("Name is not valid")
	}
	if regexp.MustCompile(`^[a-z0-9_.\-\+]+@[a-z0-9_.\-\+]+$`).MatchString(form.Get("email")) != true {
		return "email", errors.New("Email address is not valid")
	}
	if form.Get("website") != "" && regexp.MustCompile(`(https?:\/\/)?[a-z0-9\-\.]+`).MatchString(form.Get("website")) != true {
		return "website", errors.New("Website is not valid")
	}
	if regexp.MustCompile(`[a-z0-9\-]+(\/[a-z0-9\-]+)*`).MatchString(form.Get("post_id")) != true {
		return "post_id", errors.New("post_id is not valid")
	}
	if regexp.MustCompile(`[a-z]+`).MatchString(form.Get("content_type")) != true {
		return "content_type", errors.New("Content type is not valid")
	}
	if form.Get("body") == "" {
		return "body", errors.New("You forgot to write the actual comment!")
	}
	if !postExists(form.Get("post_id"), form.Get("content_type"), contentDir) {
		return "post_id", errors.New("Specified post does not exist")
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

func postExists(postID string, extension string, contentDir string) bool {
	return isValidPath(filepath.Join(contentDir, postID + "." + extension))
}

func processBody( body string ) string {
	body = strings.Replace(body, `"""`, "%quote%", -1)
	body = template.HTMLEscapeString(body)
	body = strings.Replace(body, "%quote%", ">", -1)
	return body
}

func getGravatarId(email string) string {
	hash := md5.Sum([]byte(email))
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

func buildFilename(email string, timestamp string) string {
	return timestamp + "-" + email + ".json"
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
	dirCreated := false
	// If path doesn't exist, create it
	if !isValidPath(path) {
		os.MkdirAll(path, os.ModePerm)
		dirCreated = true
	}
	fn := filepath.Join(path, filename)
	// Write file to disk
	// FIXME: Needs error handling
	ioutil.WriteFile(fn, JSON, 0600)
	if dirCreated == true && config.TouchFile != "" {
		updateChangeFile(filepath.Join( config.BaseDir, config.TouchFile))
	}
	return nil
}

func updateChangeFile(changeFile string) {
	// FIXME: Needs error handling
	ioutil.WriteFile(changeFile, []byte("."), 0600)
}

func main() {
	sourceFlag := flag.String("src", ".", "Full path to the sources dir.")
	contentFlag := flag.String("content", "content", "The content directory. Relative to source dir.")
	commentsFlag := flag.String("comments", "comments", "The directory to save comments. Relative to source dir.")
	touchFlag := flag.String("touch", "", "File to update when a new comment directory is created. Relative to source dir. Some watch scripts may require this.")
	addressFlag := flag.String("address", "", "IP address to use for incoming requests. Defaults to any address.")
	portFlag := flag.String("port", "8080", "Port to listen to for incoming requests.")
	pathFlag := flag.String("path", "/comment", "The url path that is used for comment processing.")

	flag.Parse()

	config.BaseDir = *sourceFlag
	config.ContentDir = filepath.Join(*sourceFlag, *contentFlag)
	config.CommentsDir = filepath.Join(*sourceFlag, *commentsFlag)
	config.TouchFile = filepath.Join(*sourceFlag, *touchFlag)
	
	// FIXME: Add 
	serverAddress := *addressFlag + ":" + *portFlag;

	http.HandleFunc( *pathFlag, saveComment );
	//http.HandleFunc("/new", newComment );
	http.ListenAndServe( serverAddress, nil)
}
