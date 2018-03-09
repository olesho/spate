

if ('Notification' in window) {
    var messaging = firebase.messaging();

    subscribe();

    messaging.onMessage(function(payload) {
        console.log('Message received. ', payload);
        new Notification(payload.notification.title, payload.notification);
    });
}

function subscribe() {
    messaging.requestPermission()
        .then(function () {
            messaging.getToken()
                .then(function (currentToken) {
                    if (currentToken) {
                        sendTokenToServer(currentToken);
                    } else {
                        console.warn('Unable to receive token.');
                        setTokenSentToServer(false);
                    }
                })
                .catch(function (err) {
                    console.warn('Error receiving token.', err);
                    setTokenSentToServer(false);
                });
    }).catch(function (err) {
        console.warn('Unable to get permission.', err);
    });
}

function unsubscribe() {
    window.localStorage.setItem('sentFirebaseMessagingToken', '')
    fetch(`/subscribe/firebase`, {
      method: 'POST',
      headers: new Headers({
        'Content-Type': 'application/json',
        'token': '',
        'credentials': 'include'
      })
    })
}

function sendTokenToServer(currentToken) {
    if (!isTokenSentToServer(currentToken)) {
        fetch(`/subscribe/firebase`, {
            body: JSON.stringify({}),
            cache: 'no-cache', 
            credentials: 'same-origin', 
            headers: {
                'subscription': currentToken,
                'content-type': 'application/json'
            },
            method: 'POST',
            mode: 'cors', 
            redirect: 'follow', 
            referrer: 'no-referrer', 
        })


        setTokenSentToServer(currentToken);
    }
}

function isTokenSentToServer(currentToken) {
    return false
    //return window.localStorage.getItem('sentFirebaseMessagingToken') == currentToken;
}

function setTokenSentToServer(currentToken) {
    window.localStorage.setItem(
        'sentFirebaseMessagingToken',
        currentToken ? currentToken : ''
    );
}