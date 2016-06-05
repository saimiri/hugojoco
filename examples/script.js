/**
 * An example script for sending the comment form via Ajax.
 */
var form = document.getElementById( "comment-form" )
if ( form ) {
	form.addEventListener( "submit", function( e ){
		var req = new XMLHttpRequest();
		var msgArea = document.getElementById( "comment-form-message" );
		msgArea.class = "";
		msgArea.textContent = "";
		req.onload = function( e ){
			if ( req.status === 200) {
				msgArea.textContent = "Thank you for the comment! It should be visible after you refresh the page.";
				msgArea.classList.add( "message" );
				msgArea.classList.add( "message--success" );
				form.reset();
			} else {
				msgArea.classList.add( "message" );
				msgArea.classList.add( "message--error" );
				msgArea.textContent = req.response.Message;
			}
		};
		var fields = document.querySelectorAll( ".comment-form__field" );
		var values = [];
		for ( var i = 0, j = fields.length; i < j; i++ ) {
			values.push( fields[i].name + "=" + encodeURIComponent( fields[i].value ) );
		}
		var payload = values.join( "&" );
		req.open( "POST", "//www.example.com/comment", true );
		req.responseType = "json";
		req.setRequestHeader( "Content-type", "application/x-www-form-urlencoded" )
		req.setRequestHeader( "X-Requested-With", "XMLHttpRequest" );
		req.send( payload );
		e.preventDefault();
	} );
}