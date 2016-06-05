# hugojoco -- Hugo JSON Comments

Hugo JSON Comments is a little Go program that listens to POST requests.
If they match a specific format it processes them and saves them as JSON
files into the Hugo source folder.

These files can then read using Hugo's [`readDir` function](https://gohugo.io/extras/localfiles/)
and looped through. In the loop each of them can be read with the
[`getJSON` function](https://gohugo.io/extras/datadrivencontent/). The
decoded comment objects can then used to display the comments.

See http://saimiri.io/comments-in-hugo for a more detailed explanation.

**Note:** This is still at proof-of-concept level, although I have it in
production on my own site.

## Installation

## Usage

See `hugojoco -h` for options.

## Using With nginx

If you are using nginx you can redirect requests to **hugojoco** with
`proxy_pass`:

```
location /comment {
    proxy_pass http://127.0.0.1:8080;
  }
```