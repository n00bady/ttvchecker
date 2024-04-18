## Description
A simple tool to check if your favorite streamers are live from
the command line now with an interactive TUI made with bubbletea!  

## Install
I maintain an AUR pkg [here](https://aur.archlinux.org/packages/ttvchecker)
There are also "examples" for a PKGBUILD and VoidLinux pkg in the repo.

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

## Dependencies
A terminal that support utf-8. True color preferable but lipgloss should change the colors accordingly.  
Also it depends on xdg-open to open the streams to your default browser.

## Build
`go build .`

## TODO
* Find a way to make the response parsing more reliable because sometimes some streams appear as offline
while the are live and vice versa.
* Make it possible to add a streamer by the entire url.
* Make it cross-platform.(Works only on Linux for now.)
* Implement the streamer add and delete function in the TUI.
* Resize the TUI acording to the terminal size.
* Ability to check if a certain streamer given by argument is live.
* ...
* ???

