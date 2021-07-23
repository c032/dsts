# dsts

My custom status for i3bar.

There's no configuration files, everything is decided at compile time.

I don't even use this anymore, but I also don't want to archive it just
yet.

## Compile-time requirements

*  Go compiler. See the `go.mod' file to know which version I was using
   last time I updated this repository.

## Runtime requirements

*  A running instance of MPD. Take a look at the `cmd/dsts' package to
   know which port number is expected.

## What is being displayed

*  Current date and time in the same format as Skyrim.
*  Song currently playing in MPD.

## How it works

Each `component' is a function receiving a channel as its only argument.
Each time a status is sent to that channel, a refresh is triggered.

The main program coordinates all those updates, and sends new messages
when necessary.

This allows components to each have their own `refresh rate', instead of
using a single global refresh rate for everything.
