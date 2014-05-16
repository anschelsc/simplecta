$(function() {
	$(".ajax_link").click(function() {
		var button = $(this);
		var url;
		var mark = button.data("mark")
		if (mark === "read") {
			url = "/markRead/";
		} else {
			url = "/markUnread/";
		}
		$.get(url, button.data("key"), function() {
			if (mark === "read") {
				mark = "unread";
			} else {
				mark = "read";
			}
			button.text("mark " + mark);
			button.data("mark", mark)
		});
	});
	$(".read_link").bind("mouseup", function() {
		var button = $(this).siblings("button");
		if (button.data("mark") === "read") {
			button.text("mark unread");
			button.data("mark", "unread");
		}
	});
});