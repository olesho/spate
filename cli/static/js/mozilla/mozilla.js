if (isFirefox) {
		
  if ('serviceWorker' in navigator && 'PushManager' in window) {
    console.log('Service Worker and Push is supported');

    navigator.serviceWorker.ready.then(function(reg) {
      console.log("ready")

    })

    navigator.serviceWorker.register('/js/mozilla/sw.js')
    .then(function(reg) {
      reg.update().then(() => {
        reg.pushManager.subscribe().then(sub => {
          fetch(`/subscribe/mozilla`, {
            method: 'POST', // or 'PUT'
            body: JSON.stringify(sub.toJSON()), 
            headers: new Headers({
              'Content-Type': 'application/json'
            })
          })
        })
      })
    })
    .catch(function(error) {
      console.error('Service Worker Error', error);
    });
   
  } else {
    console.warn('Push messaging is not supported');
    pushButton.textContent = 'Push Not Supported';
  }
  
}