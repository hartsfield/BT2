var tracklist = [];
var nowPlaying = {                                                                                                      
        "isPlaying": false,                                                                                             
        "hasPlayed": false,                                                                                             
        "ID": undefined,                                                                                                
        "path": undefined,
        "key": undefined,
};                                                                                                                      
                                                                                                                        
// Pause/play (pp()): Changes the pause/play icon to reflect the state of the                                           
// track                                                                                                                
var track                                                                                                               
function pp(trackID, trackPath, key) { 
        if (key != undefined) {
                nowPlaying.key = key;
        }
console.log(tracklist[key])
        if (tracklist[key].liked) {
                glb = document.getElementById("globalLikeButt");
                glb.style.backgroundImage = "url(/public/assets/heart_red.svg)";
        } else {
                glb = document.getElementById("globalLikeButt");
                glb.style.backgroundImage = "url(/public/assets/heart_black.svg)";
        }
        if (nowPlaying.ID != trackID && trackID != undefined) {                                                         
                if (nowPlaying.ID != undefined) {                                                                       
                        track.pause();                                                                                  
                        track.src = "../"+ trackPath;                                                                   
                }                                                                                                       
                track = document.getElementById("audioID_global");                                                      
                track.addEventListener('timeupdate', (event) => {                                                       
                                tt()                                                                                    
                });                                                                                                     
        }                                                                                                               
                                                                                                                        
        if (trackID == undefined && trackPath == undefined) {                                                           
                trackID = nowPlaying.ID;                                                                                
        } else if (nowPlaying.path != trackPath){                                                                       
                track.src = "../" + trackPath;                                                                          
                nowPlaying.path = trackPath;                                                                            
        }                                                                                                               
        if (track.paused) {                                                                                             
                track.play();                                                                                           
                nowPlaying.ID = trackID;                                                                                
                document.getElementById("ppImg_global").style.backgroundImage = "url(/public/assets/pause.png)";       
                document.getElementById("ppButt_" + trackID).style.backgroundImage = "url(/public/assets/pause.png)";                            
        } else {                                                                                                        
                track.pause();                                                                                          
                document.getElementById("ppImg_global").style.backgroundImage = "url(/public/assets/play.png)"; 
                document.getElementById("ppButt_" + trackID).style.backgroundImage = "url(/public/assets/play.png)";                      
        }                                                                                                               
}                                                                                                                       

// Time tracker (tt()): Runs ontimeupdate and expands the innerSeeker element on
// the global player to reflect the time position of the audio track
function tt() {
        document.getElementById('innerSeeker').style.width = (Math.floor(track.currentTime) /
                Math.floor(track.duration)) * 100 + "%";
        // document.getElementById("spinny").style.transform = "rotate("+Math.floor(track.currentTime)*5+"deg)";
}

// seek() gets the mouses x-cooridinate when it clicks the outerSeeker div and                                 
// uses this information to seek to a relative position in the audio track                                     
function seek(e) {                                                              
        var sizer = document.getElementById("outerSeeker");                                                    
        var playButt = document.getElementById("ppImg_global");                                                
        var seekTo = ((track.duration / 100) * ((e.clientX - sizer.offsetLeft - (window.innerWidth - sizer.offsetLeft - sizer.offsetWidth - playButt.offsetWidth)) / sizer.offsetWidth) * 100);
        track.currentTime = seekTo;                                                                            
}

// function trackLink(id) {
//   window.location = window.location.origin+'/track/'+id;
// }

// Listens for when a user chooses a new song and changes all icons in the music                                          
// list to a play button except for the icon associated with the chosen song.                                             
// Also adds nowPlaying information to the global "nowPlaying" object, and                                                
// updates the info in the global player.                                                                                 
document.addEventListener('play', function(e) {                                                                           
                document.getElementById("controls").style.display = "unset";                                              
                if (nowPlaying.hasPlayed == false) {                                                                      
                        var c1 = document.getElementById("ppImg_global").offsetWidth;                                     
                        var c2 = document.getElementById("globalNextButt").offsetWidth;                                   
                        var c3 = document.getElementById("globalLikeButt").offsetWidth;                                   
                        // document.getElementById("outerSeeker").style.width = (window.innerWidth - (sb+c1+c2+c3)) + "px";  
                        document.getElementById("outerSeeker").style.marginLeft= (c1+c2+c3) + "px";                          
                        nowPlaying.hasPlayed = true;                                                                      
                }                                                                                                         
                updateTrackList();                                                                                        
}, true);                                                                                                                 
                                                                                                                          
function updateTrackList() {                                                                                              
        var audios = document.getElementsByClassName('post');                                                          
        track = document.getElementById("audioID_global");                                                                
        for (var i = 0, len = audios.length; i < len; i++) {                                                              
                var pp = document.getElementById("ppButt_" + audios[i].id);                                                
                if (track.paused == true || audios[i].id != nowPlaying.ID) {                                              
                        pp.style.backgroundImage = "url(/public/assets/play.png)";                                                         
                } else {                                                                                                  
                        nowPlaying.isPlaying = true;                                                                      
                        nowPlaying.artist = audios[i].dataset.artist;                                                     
                        nowPlaying.title = audios[i].dataset.title;                                                       
                        document.getElementById("globalTrackInfo").innerHTML = nowPlaying.artist 
                          + " - " + nowPlaying.title;                                    
                        // pp.src = "public/assets/pause.png";                                                        
                        pp.style.backgroundImage = "url(/public/assets/pause.png)";                            
                }                                                                                                         
        }                                                                                                                 
}                                                                                                                         

function getStream(category, tID) {
        var xhr = new XMLHttpRequest();

        xhr.open("POST", "/api/getStream");
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.onload = function() {
                if (xhr.status === 200) {
                        var res = JSON.parse(xhr.responseText);
                        if (res.success == "true") {
                                tracklist = JSON.parse(res.tracklist);
                                var listDiv = document.getElementById("updateDiv");
                                listDiv.innerHTML = res.template;
                                updateTrackList();
                                if (category == 'TRACK') {
                                        window.history.pushState({}, "page", "/track/" + tID);
                                        window.scrollTo(0, 0);
                                } else {
                                        window.history.pushState({}, "page", "/#/" + category);
                                        window.scrollTo(0, 0);
                                }
                                // listDiv.insertAdjacentHTML("beforeend", res.template);
                        } else {
                                // handle error
                        }
                }
        };

        // For now, all we're sending is a username and password, but we may start
        // asking for email or mobile number at some point.
        xhr.send(JSON.stringify({
                category: category,
                pageNumber: tID
        }));
        // var sb = document.getElementById("sidebar").offsetWidth;
        // var s = document.getElementById("sizer");
        // w = (window.innerWidth - (sb)) + "px";
        // mL = (sb) + "px";
        // s.style.width = w;
        // s.style.marginLeft = mL;
}

function like(trackID, isLoggedIn, key) {
        // if (isLoggedIn == "false") {
        //         showAuth();
        // } else {
                var xhr = new XMLHttpRequest();

                xhr.open("POST", "/api/like");
                xhr.setRequestHeader("Content-Type", "application/json");
                xhr.onload = function() {
                        if (xhr.status === 200) {
                                var res = JSON.parse(xhr.responseText);
                                if (res.success == "false") {
                                        // If we aren't successful we display an error.
                                        document.getElementById("errorField").innerHTML = res.error;
                                } else if (res.isLiked == "true") {
                                        document.getElementById("heart_" + trackID).style.backgroundImage = "url(/public/assets/heart_red.svg)";
                                        tracklist[key].liked = true;
                                        if (nowPlaying.ID == trackID) {
                                                glb = document.getElementById("globalLikeButt");
                                                glb.style.backgroundImage = "url(/public/assets/heart_red.svg)";
                                        }

                                } else if (res.isLiked == "false") {
                                        document.getElementById("heart_" + trackID).style.backgroundImage = "url(/public/assets/heart_black.svg)";
                                        tracklist[key].liked = false;
                                        if (nowPlaying.ID == trackID) {
                                                glb = document.getElementById("globalLikeButt");
                                                glb.style.backgroundImage = "url(/public/assets/heart_black.svg)";
                                        }
                                } else {
                                        // handle error
                                        console.log("error");
                                }
                        }
                };

                // For now, all we're sending is a username and password, but we may start
                // asking for email or mobile number at some point.
                xhr.send(JSON.stringify({
                                        id: trackID
                }));

        // }
}

function nextTrack() {
  nt = tracklist[nowPlaying.key + 1];
  pp(nt.ID, nt.path, nowPlaying.key + 1);
}

function globalLike(isLoggedIn) {
   like(nowPlaying.ID, isLoggedIn);
}
