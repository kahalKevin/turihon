new Vue({
    el: '#app',
    data: {
        ws: null, // Our websocket
        newMsg: '', // Holds new messages to be sent to the server
        own_address: '', // Holds our own address
        to_address: '', //Where will the chat sent
        chatContent: {}, // A running list of chat messages displayed on the screen
        chatContentOne: '',
        contact_list: [],
        email: null, // Email address used for grabbing an avatar
        username: null, // Our username
        joined: false // True if email and username have been filled in
    },

    created: function() {
        var self = this;
        var own_address;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            console.log(msg);
            if (msg.type == 'message'){
                if (typeof self.chatContent[msg.from] == 'undefined'){
                    self.chatContent[msg.from] = new String('');
                }
                self.chatContent[msg.from] += '<div class="chip">'
                        + '<img src="' + self.gravatarURL(msg.email) + '">' // Avatar
                        + msg.username
                    + '</div>'
                    + emojione.toImage(msg.message) + '<br/>'; // Parse emojis

                var element = document.getElementById('chat-messages');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
            }
            if (msg.type == 'init'){
                self.chatContent = {};
                self.own_address = new String(msg.from);
                self.contact_list = msg.contact;
            }
            if (msg.type == 'add'){
                if (self.contact_list == null ){
                    self.contact_list= [];
                }
                self.contact_list.push(msg.from);
            }
            if (msg.type == 'remove'){
                console.log(msg.from);
                // Remove disconnected user from contact list
                var index = self.contact_list.indexOf(msg.from);
                self.contact_list.splice(index,1);

                // Delete conversation done with disconnected user
                delete self.chatContent[msg.from];
            }            
        });
    },

    methods: {
        send: function () {
            if (this.newMsg != ''  &&  this.to_address != '') {
                this.ws.send(
                    JSON.stringify({
                        to_address: this.to_address,
                        from: this.own_address,
                        type : 'message',
                        email: this.email,
                        username: this.username,
                        message: $('<p>').html(this.newMsg).text() // Strip out html
                    }
                ));
                // Add this to our chat window too
                this.chatContent[this.to_address] += '<div class="chip">'
                        + '<img src="' + this.gravatarURL(this.email) + '">' // Avatar
                        + this.username
                    + '</div>'
                    + emojione.toImage( $('<p>').html(this.newMsg).text() ) + '<br/>'; // Parse emojis

                var element = document.getElementById('chat-messages');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom

                this.newMsg = ''; // Reset newMsg
                this.to_address = ''; // Reser to_address
            }
        },

        join: function () {
            if (!this.email) {
                Materialize.toast('You must enter an email', 2000);
                return
            }
            if (!this.username) {
                Materialize.toast('You must choose a username', 2000);
                return
            }
            this.email = $('<p>').html(this.email).text();
            this.username = $('<p>').html(this.username).text();
            this.joined = true;
        },

        selectContact : function (selected){
            this.chatContentOne = this.chatContent[selected];
            this.to_address = selected;
        },

        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + CryptoJS.MD5(email);
        }
    }
});