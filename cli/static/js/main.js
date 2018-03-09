jQuery(document).ready(function() {
    jQuery(".block-request-body").hide()

    jQuery("#request_method").change(function() {
        if (jQuery("option:selected", "#request_method").val() !== "GET") {
            jQuery(".block-request-body").show()  
            return
        }
        jQuery(".block-request-body").hide()     
    })

    jQuery.get('/endpoint', (data) => {
            var data = JSON.parse(data)
            jQuery("#request_interval").val(data.min_interval)
            jQuery("#request_method").val(data.method)
            jQuery("#request_url").val(data.url)
            jQuery("#request_headers").val(data.header)
            jQuery("#request_body").val(data.body)
    })

    jQuery('#request_submit').click(e => {
        var endpoint = {
            "min_interval": parseInt(jQuery("#request_interval").val()),
            "method": jQuery("#request_method").find(":selected").val(),
            "url": jQuery("#request_url").val(),
            "header": jQuery("#request_headers").val(),
            "body": jQuery("#request_body").val(),
        }
        jQuery.ajax(`/endpoint`, {
                    data : JSON.stringify(endpoint),
                    contentType : 'application/json',
                    type : 'POST'
        }).done((data) => {
                    console.log(data)
        })
    })

    jQuery('#request_subscription_unsubscribe').click(e => {
        unsubscribe()
    })
})

