self.addEventListener('push', function(event) {
  var data = JSON.parse(event.data.text())
  const title = data.Title;
  const options = {
    body: data.Body,
    actions: [
    	{action: "get", title: "Get now."},
    	{action: "click", title: data.Url},
    ],
    //icon: 'images/icon.png',
    //badge: 'images/badge.png'
    tag: data.Url
  };

  event.waitUntil(self.registration.showNotification(title, options));
});


self.addEventListener('notificationclick', function(event) {
	event.notification.close()
	console.log("Event action", event.notification.tag)
	event.waitUntil(self.clients.openWindow(event.notification.tag));
})
