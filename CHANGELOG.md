# Changelog

This file lists the main changes with each version of the FyneDesk project.
More detailed release notes can be found on the [releases page](https://github.com/fyne-io/desktop/releases). 

## 0.2 - 22 March 2021

### Added

* Support for desktop notifications
* Created API for writing modules that plug in to FyneDesk
* Add keyboard shortcut support for modules
* Add volume control using pulseaudio
* Print screen support for desktop and window capture
* Crash logs are now saved before runner restarts the desktop
* Add urls, calculations and sound/brightness to launcher
* Add support for XPM icon format
* Double tap title bar to maximise
* Hover effects for window borders
* Add option to use 24h time
* Add support for "fake" transparency in X apps
* Drag border to exit un-maximize
* Support choosing between Alt and Win key for shortcuts
* Add AC power indicator to battery module
* Add option to change border button position

### Changed

* Updated to latest Fyne theme
* Updated multi-monitor support
* Bar icons can be added and removed from a new right-click menu
* Move to makefile for system installation
* Search user's local directory for applications
* New Swedish Pine wallpaper by @Jacalz
* Better support for running in virtual machines
* Improved background selection for settings
* Account menu now appears over windows

### Fixed

* Communicate the desktop DPI for better support in apps not using GTK, Qt or EFL
* When starting with multiple screens that are not configured they overlap strangely (#129)
* Update scaling per-screen for Qt apps
* Drag and drop in Chrome browser (#156)
* Firefox does not enter Fullscreen (#110)
* Icon bar flicker when hovering
* Add BSD support for all status modules
* Fix 12h time format (#145)
* Killing Xorg with fynedesk_runner will not exit cleanly (#137)
* Menu opens with 0 height (#114)
* Allow resize from top corners (#113)
* Windows can be slow to move around (#165)
* Crash on fast alt-tab (#122)
* alt-space not reliable (#160)
* Moving windows on external screens shows screen tearing (#161)
* Applications can push other ones off the screen (#163)


## 0.1.3 - 15 April 2020

Renamed package to fyne.io/fynedesk and repository to github.com/fyne-io/fynedesk.


## 0.1.2 - 13 April 2020

Additional bug fixes and graphical tweaks on 0.1 release

### Added

* Initial support for BSD systems

### Changed

* Improved wording for account menu in embedded mode
* Simpler app bar configuration screen
* Updated default background for a winter scene

### Fixed

* GoLand editor loses focus on mouse click (#69)
* Fix a flicker on window borders when uncovered (#83)
* Respect the min and max hints of windows (#85)
* Update to latest Fyne release to match new scale calculations
* Drag and drop targets not found for many applications (#49)
* VLC opens a lot of windows (#70)
* Graphical glitches when Qt apps when scale < 1.0
* Improve loading of macOS apps and icons in embedded mode on darwin


## 0.1.1 - 30 December 2019

Bug fix release on 0.1

### Features

* Added key bindings for app launcher (open with Alt-Space)
* Add an about screen for FyneDesk

### Fixed

* Issue where application menus may not be aligned in maximised or full screen
* Fix crash if right clicking app launcher input field
* Correct issue where icon list may not be saved from settings


## 0.1 - 23 December 2019

This first release introduces the FyneDesk project to the Linux community.
FyneDesk is a full X11 window manager that can be used in place of software like
Gnome or KDE. Using the Fyne project means that it uses scalable vector graphics
throughout whilst providing excellent performance and stability.

### Features

* X11 window management and window borders
* Alt-Tab stack cycling for switching applications
* Application launcher
* Task and launch icon bar
* Battery and screen brightness controls
* Light and Dark themes
* Multiple monitor support

