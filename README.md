## Description
A simple tool to check if your favorite streamers are live without using the twitch api from
the command line now with an interactive TUI made with bubbletea!  

## Install
I maintain an AUR pkg [here](https://aur.archlinux.org/packages/ttvchecker) just install it 
with `yay -S ttvchecker` or whatever aur manager you use.

## Usage
`ttvchecker add <streamerName>` to add a streamer.  
`ttvchecker del <streamerName>` to delete a streamer.  
You can add/del several at the same time by seperating them with a space you can either just add 
their username or the full url to their stream, either should work.
Of course you can also add the streamers directly to the config file in `$HOME/.config/ttvchecker/`
`ttvchecker check` to see which streamers are live.  
`ttvchecker tui` to start the interactive TUI!  


For more see `ttvchecker help`

## Dependencies
A terminal that support utf-8. True color preferable but lipgloss should change the colors accordingly.  
Also it depends on xdg-open to open the streams to your default browser.

## Build
`go build .`

## TODO
* Maybe make it cross-platform ? (Works only on Linux for now.)
* Implement the streamer add and delete function in the TUI.
* Ability to check if a certain streamer is live given by argument.
* ...
* ???

