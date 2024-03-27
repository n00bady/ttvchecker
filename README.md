## Description
A simple tool to check if your favorite streamers are live from
the command line.  
Essentially my [twitchlivechecker](https://gitlab.com/n00bady/twitchlivechecker) but better and written in GO.  
Also it includes a TUI that is made with bubbletea.

## Build
`go build .`

## Usage
`ttvchecker add <streamerName>` to add a streamer.  
`ttvchecker del <streamerName>` to delete a streamer.  
You can add/del several at the same time by seperating them with a space.
`ttvchecker check` to see which streamers are live.  
`ttvchecker tui` to start the interactive TUI!

You need to add the streamer's name as is being shown in the url e.g.:  
  To add [KEYB1ND's](https://www.twitch.tv/keyb1nd_) stream you need to go
  to his stream and use the name that appears in the url https://www.twitch.tv/`keyb1nd_` <- this one
  not the one that shown in the page or the title, because they can be different.
Of course you can also add the streamers directly to the config file in `$HOME/.config/ttvchecker/`

For more see `ttvchecker help`

## TODO
* Find a way to make the response parsing more reliable because sometimes some streams get parsed as offline
while the are live.
* Make it possible to add a streamer by the entire url.
* Make it cross-platform.(Currently only works well in Linux OS)
* ...
* ???

