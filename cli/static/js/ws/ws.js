if (!isFirefox) {
    function onPermissionGranted () {
        console.log('Permission has been granted by the user');
        listen()
    }
    function onPermissionDenied () {
        console.warn('Permission has been denied by the user');
    }

    function onNotifyShow() {
            //console.log('notification was shown!');
    }

    function listen() {
            const socket = new WebSocket('ws://ai2ai/ws');
            // Listen for messages
            socket.addEventListener('message', function (event) {
                    var msg = JSON.parse(event.data)

                    var myNotification = new Notify(msg.Title, {
                            body: msg.Data,
                            notifyClick: () => {
                                    window.open(msg.Url, '_blank');
                            },
                            notifyShow: onNotifyShow
                    });

                    myNotification.show();
                    //console.log('Message from server ', event.data);
            });
    }

    if (!Notify.needsPermission) {        
            listen()
    } else if (Notify.isSupported()) {
        Notify.requestPermission(onPermissionGranted, onPermissionDenied);
    }
}