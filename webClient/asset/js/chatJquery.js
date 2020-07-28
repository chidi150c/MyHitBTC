   $(function(){ 
        var socket = null; 
        var msgBox = $("#chatmessagebox textarea");
        var user = $("#udp").text();
        var senderbefore = "";
            $("#chatmessagebox").submit(function(){ 
                if (!msgBox.val()) return false; 
                if (!socket) { 
                    alert("Error: There is no socket connection."); 
                    return false; 
                } 
                // msgBox.val()  
                socket.send(JSON.stringify({sender: user, receiver: "all", msg: msgBox.val(), role: "notadmin", channel: "public"})); 
                msgBox.val(""); 
                return false; 
            }); 
            if (!window["WebSocket"]) { 
                alert("Error: Your browser does not support web  sockets.") 
            } else { 
                socket = new WebSocket("ws://"+ window.location.host +"/chat/ws"); 
              
                socket.onclose = function() { 
                    alert("Connection has been closed."); 
                } 
                socket.onmessage = function(e) { 
                    var msg = JSON.parse(e.data);
                    if (msg.role == "Admin"){
                        // staffmsg(msg.data)
                        if (senderbefore == msg.sender){
                        $("#cchat").append(" <div class='chat friend'><ul class='chat-message' id='messages'>"+msg.msg+"</ul></div> ");
                        }else{
                        $("#cchat").append(" <div class='chat friend'><div class='w3-text-black'><img class='user-photo' src='/tools/asset/img/maturewoman.png'>"+msg.sender+"</div><ul class='chat-message' id='messages'>"+msg.msg+"</ul></div> ");     
                        }
                    }else {
                        // usermsg(msg.dat)
                        if (senderbefore == msg.sender){
                        $("#cchat").append(" <div class='chat self'><ul class='chat-message' id='messages'>"+msg.msg+"</ul></div> ");
                    }else{
                        $("#cchat").append(" <div class='chat self'><div class='w3-text-black'><img class='user-photo' src='/tools/asset/img/man.png'>"+msg.sender+"</div><ul class='chat-message' id='messages'>"+msg.msg+"</ul></div> ");    
                    }
                    }
                    senderbefore = msg.sender
                } 
            } 
    }); 
    