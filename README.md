# Games With Go

Following a video series that teaches programming via a series of small game related projects.
Start here https://www.youtube.com/watch?v=9D4yH7e_ea8.

![Current progress screenshot](https://i.imgur.com/4rzAxno.png)

## Why Go?

- Garbage collection is very low latency, compared to Java or C#, which focus on throughput. Helps reduces problems with pauses during gc.
- Compiles to native executables. Simplifies sharing games. Improved performance because it does not rely on Just-in-time compilation

## Prerequisites

What things you need to install the software and how to install them.

- Install Go and Visual Studio Code.
- Install the Go extension in VSCode. It will generate everything else.

## Getting Started With SDL2

This guide is based on https://github.com/veandco/go-sdl2.

1. Install GCC for Windows
Binding SDL2 in Go requires Windows to compile C code.
http://mingw-w64.org/doku.php/download/mingw-builds

- Version: latest
- Architecture: x86_64
- Threads: win32
- Exception: seh
- Build revision: latest

2. Download SDL2 development library (MinGW 32/64-bit). https://www.libsdl.org/download-2.0.php

3. Drag the contents of `SDL2-2.0.8\x86_64-w64-mingw32` into `C:\Program Files\mingw-w64\mingw64\x86_64-w64-mingw32`

4. Set Environment variables
Add these as new lines to the System Path variable:
- For SDL2: > `C:\Program Files\mingw-w64\x86_64-8.1.0-win32-seh-rt_v6-rev0\mingw64\x86_64-w64-mingw32\bin`
- For GCC: > `C:\Program Files\mingw-w64\x86_64-8.1.0-win32-seh-rt_v6-rev0\mingw64\bin`

5. Get SDL2 binding for Go
- `$ go get -u github.com/veandco/go-sdl2/sdl`

6. Download SDL2_ttf development library (MINGW 32/64-bit). https://www.libsdl.org/projects/SDL_ttf/
- Copy the contents of `bin`, `include/SDL2` and `lib` into their respective folders at `C:\Program Files\mingw-w64\x86_64-8.1.0-win32-seh-rt_v6-rev0\mingw64\x86_64-w64-mingw32`

5. Get SDL2_ttf binding for Go
- `$ go get -v github.com/veandco/go-sdl2/ttf`

## Help

If you encounter exit code: 3221225781, simply restart your computer.