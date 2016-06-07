# Hugo JSON Comments

Hugo JSON Comments is a little Go program that listens to POST requests. If they match a specific format it processes them and saves them as JSON files into the Hugo source folder.

These files can then read using Hugo's [`readDir` function](https://gohugo.io/extras/localfiles/) and looped through. In the loop each of them can be read with the [`getJSON` function](https://gohugo.io/extras/datadrivencontent/). The decoded comment objects can then used to display the comments.

See http://saimiri.io/comments-in-hugo for a more detailed explanation.

**Note:** This is still at proof-of-concept level, although I have it in production on my own site.

## Installation

## Usage

Typical usage might be:

``` shell-session
$ hugojoco -source=/path/to/hugo -touch=.comment -salt=5VpsqeMX6N0IiUw3e0zfPLxDQ9J5CScvW0nQhUUWNfwziPDeDKHLA60LCsmUcsL2jfmIcChZXtnv1NhGOhpRsQ6o9OyUyeU3ZzDBlD6FTGOLInkm8dia3NuaSsPwlct4
```

This launches hugojoco, which then listens for incoming POST requests until terminated.

With default settings, hugojoco 
  - processes comments posted at `:8080/comment`
  - saves each comment to `./comments` directory
  - tries to look for content files in `./content` directory
  - uses no salt for hashing email addresses

## Switches

See `hugojoco -h` for options.

### -source

Sets the source directory of your site. **Default:** "."

### -content

Sets the content directory of your site _relative to the source directory_. **Default:** "content"

### -comments

Sets the comments directory (that is, where the comments are saved) of your site _relative to the source directory_. **Default:** "comments"

### -salt

The salt to be used when hashing the email addresses for internal use. This is not intended to make the hash more secure but to make it unique for each site. **Default:** none

### -address

The IP address to use for the server. **Default:** any address

### -port

The port that the server listens to. **Default:** 8080

### -path

The URL path that is used to process comments. **Default:** /comment

## Using With nginx

If you are using nginx you can redirect requests to hugojoco with `proxy_pass`:

```
location /comment {
    proxy_pass http://127.0.0.1:8080;
  }
```