$(function() {
	var client = $("head").data("client");
	var setup = function() {
		$.getJSON("/getToken/", {client: client}, function(token) {
			var chan = new goog.appengine.Channel(token);
			var sock = chan.open();
			sock.onmessage = function(m) {
				var res = $.parseJSON(m.data);
				var button = $("button[data-index="+res.Ind+"]");
				var mark;
				if (res.Read) {
					mark = "unread";
				} else {
					mark = "read";
				}
				button.text("mark " + mark);
				button.data("mark", mark);
			};
			//sock.onclose = setup;
		});
	};
	setup();

	$(".ajax_link").click(function() {
		var button = $(this);
		var url;
		var mark = button.data("mark")
		if (mark === "read") {
			url = "/markRead/";
		} else {
			url = "/markUnread/";
		}
		$.get(url, button.data());
	});
});