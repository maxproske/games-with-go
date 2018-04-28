# Games With Go

Following a video series that teaches programming via a series of small game related projects.
Start here https://www.youtube.com/watch?v=9D4yH7e_ea8.

## Prerequisites

What things you need to install the software and how to install them.

Install golang and Visual Studio Code.

Install the go extension. It will generate everything else.

## Getting Started With SDL2

Follow the steps here: https://github.com/veandco/go-sdl2

1. Install GCC for Windows
Binding SDL2 in Go requires Windows to compile C code.
http://mingw-w64.org/doku.php/download/mingw-builds (During installation, choose x86_64)

Version: latest
Architecture: x86_64
Threads: win32
Exception: seh
Build revision: 1

2. Download SDL2 (MinGW 32/64-bit)
https://www.libsdl.org/download-2.0.php

3. Drag the contents of SDL2-2.0.8\x86_64-w64-mingw32 into C:\Program Files\mingw-w64\mingw64\x86_64-w64-mingw32

4. Set Environment variables
Add these as new lines to the System Path variable:
For SDL2: > C:\Program Files\mingw-w64\x86_64-7.3.0-posix-seh-rt_v5-rev0\mingw64\x86_64-w64-mingw32\bin
For GCC: > C:\Program Files\mingw-w64\x86_64-7.3.0-posix-seh-rt_v5-rev0\mingw64\bin

## Help

If you encounter exit code: 3221225781, simply restart your computer.