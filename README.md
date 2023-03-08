## Description
A simple tool to check if your favorite streamers are live from
the command line.  
Essentially my [twitchlivechecker](https://gitlab.com/n00bady/twitchlivechecker) but better and written in GO.  

## Build
`go build`

## Usage
`ttvchecker add <streamerName>` to add a streamer (you can add multiple).  
`ttvchecker del <streamerName>` to delete a streamer (you can del multiple).  
`ttvchecker check` to see which streamers are live.  

You need to add the streamer's name as is being shown in the url e.g.:  
  To add [KEYB1ND's](https://www.twitch.tv/keyb1nd_) stream you need to go
  to his stream and use the name that appears in the url https://www.twitch.tv/`keyb1nd_` <- this one
  not the one that shown in the page or the title, because they can be different.

## TODO
* Find a way to make the response parsing more reliable because sometimes some streams get parsed as offline
while the are Live.
* Make it possible to add a streamer by the entire url.
* ...
* ???

