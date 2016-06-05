# Hugo JSON Comments

Hugo JSON Comments is a little program that listens to POST requests and
if they match a specific format it processes them and saves them as JSON
files into your Hugo source folder.

These files can then read using Hugo's [`readDir` function](https://gohugo.io/extras/localfiles/)
and looped through. In the loop each of them can be read with the
[`getJSON` function](https://gohugo.io/extras/datadrivencontent/). The
decoded comment object can then used to display the comment.

See http://saimiri.io/comments-in-hugo for a more detailed explanation.

## Usage

See `hugojoco -h` for options.