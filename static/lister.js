$(function() {
	$(".ajax_link").click(function() {
		var button = $(this);
		if (button.html() === "show feed URL") {
			button.next().removeAttr("hidden");
			button.html("hide feed URL");
		} else {
			button.next().attr("hidden", "true");
			button.html("show feed URL");
		}
	});
});