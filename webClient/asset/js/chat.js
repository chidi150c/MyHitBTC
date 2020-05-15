function WebSocketTest(simid) {
   if ("WebSocket" in window) {      
      // Let us open a web socket
      var ws = new WebSocket("ws://"+ window.location.host +"/feeds/ws");
      ws.onopen = function() {         
         simid = simid.replace(/['"]+/g, '')
         ws.send(JSON.stringify({symbolcode: simid}));
        };      
      ws.onmessage = function (evt) {
        var rdata = JSON.parse(evt.data);
    var dat = `<div class="w3-container w3-cell w3-mobile">
      <div class="w3-row-padding" style="margin:0 -16px">
          <div class="w3-twothird" >       
            <div class="w3-bar w3-border w3-blue">
              <div class="w3-bar-item">Feeds for `+rdata.symbolcode+`</div>
            </div>
            <table class="w3-table w3-striped w3-white">
              <tr>
                <td><i class="fa fa-user w3-text-blue w3-large"></i></td>
                <td>Symbol Code</td>
                <td><i>`+rdata.symbolcode+`</i></td>
              </tr>
              <tr>
                <td><i class="fa fa-user w3-text-blue w3-large"></i></td>
                <td>Successful Orders</td>
                <td><i>`+rdata.successfulorders+`</i></td>
              </tr>
              <tr>
                <td><i class="fa fa-user w3-text-blue w3-large"></i></td>
                <td>Made Profit Orders</td>
                <td><i>`+rdata.madeprofitorders+`</i></td>
              </tr>
              <tr>
                <td><i class="fa fa-bookmark w3-text-blue w3-large"></i></td>
                <td>Made Lost Orders</td>
                <td><i>`+rdata.madelostorders+`</i></td>
              </tr>           
              <tr>
                <td><i class="fa fa-share-alt w3-text-green w3-large"></i></td>
                <td>Trail Points</td>
                <td><i>`+rdata.trailpoints+`</i></td>
              </tr>
              <tr>
                <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
                <td>Good Biz</td>
                <td><i>`+rdata.goodbiz+`</i></td>
              </tr>
              <tr>
                <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
                <td>Least Profit Margin</td>
                <td><i>`+rdata.leastprofitmargin+`</i></td>
              </tr>
              <tr>
                <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
                <td>Hodling</td>
                <td><i>`+rdata.hodler+`</i></td>
              </tr>
            </table>
          </div>
      </div>
    </div>
    <div class="w3-container w3-cell w3-mobile" >
      <div class="w3-row-padding" style="margin:0 -16px">
        <div class="w3-twothird" >       
          <div class="w3-bar w3-border w3-blue" style="width:100%">
            <div class="w3-bar-item">Feeds for `+rdata.symbolcode+`</div>
          </div>
          <table class="w3-table w3-striped w3-white">
            <tr>
              <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
              <td>Net Profit or Lost</td>
              <td><i>`+ String(parseFloat(rdata.totalprofit) + parseFloat(rdata.totallost)) +`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
              <td>Instant Profit</td>
              <td><i>`+rdata.instantprofit+`</i></td>
            </tr> 
            <tr>
              <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
              <td>Total Profit</td>
              <td><i>`+rdata.totalprofit+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-comment w3-text-red w3-large"></i></td>
              <td>Instant Lost</td>
              <td><i>`+rdata.instantlost+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-users w3-text-yellow w3-large"></i></td>
              <td>Total Lost</td>
              <td><i>`+rdata.totallost+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-laptop w3-text-red w3-large"></i></td>
              <td>Never Bought</td>
              <td><i>`+rdata.neverbought+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-laptop w3-text-red w3-large"></i></td>
              <td>Never Sold</td>
              <td><i>`+rdata.neversold+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-laptop w3-text-red w3-large"></i></td>
              <td>Disable Transaction</td>
              <td><i>`+rdata.disabletransaction+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-comment w3-text-red w3-large"></i></td>
              <td>PendingA Slice</td>
              <td><i>`+rdata.pendinga+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-comment w3-text-red w3-large"></i></td>
              <td>PendingB Slice</td>
              <td><i>`+rdata.pendingb+`</i></td>
            </tr>
          </table>
        </div>
      </div>
    </div>
    <div class="w3-container w3-cell w3-mobile" >
      <div class="w3-row-padding" style="margin:0 -16px">
        <div class="w3-twothird" >       
          <div class="w3-bar w3-border w3-blue" style="width:100%">
            <div class="w3-bar-item">Feeds for `+rdata.symbolcode+`</div>
            <div class="w3-bar-item w3-left"><a href="/resetapp?appid=`+rdata.symbolcode+`">Reset</a></div>
            <div class="w3-bar-item w3-left"><a href="/editapp?symbolcode=`+rdata.symbolcode+`">Edit</a></div>
            <div class="w3-bar-item w3-left"><a href="/close?appid=`+rdata.symbolcode+`">Close</a></div>
            <div class="w3-bar-item w3-right"><a href="/deleteapp?appid=`+rdata.symbolcode+`">Delete</a></div>
          </div>
          <table class="w3-table w3-striped w3-white">
            <tr>
              <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
              <td>Message</td>
              <td><i>`+rdata.message+`</i></td>
            </tr>           
            <tr>
              <td><i class="fa fa-bell w3-text-red w3-large"></i></td>
              <td>Quantity Increment</td>
              <td><i>`+rdata.quantityincrement+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-users w3-text-yellow w3-large"></i></td>
              <td>Log Message Filter</td>
              <td><i>`+rdata.messagefilter+`</i></td>
            </tr>          
            <tr>
              <td><i class="fa fa-laptop w3-text-red w3-large"></i></td>
              <td>Stoplost Selling Point</td>
              <td><i>`+rdata.stoplostpoint+`</i></td>
            </tr>
            <tr>
              <td><i class="fa fa-laptop w3-text-red w3-large"></i></td>
              <td>Sure Trade Factor</td>
              <td><i>`+rdata.suretradefactor+`</i></td>
            </tr>
          </table>
        </div>
      </div>
    </div>`

         document.getElementById('feeder').innerHTML = dat;       
      };
      ws.onclose = function() {          
         // websocket is closed.
      };
   } else {     
      // The browser doesn't support WebSocket
      alert("WebSocket NOT supported by your Browser!");
   }
}

