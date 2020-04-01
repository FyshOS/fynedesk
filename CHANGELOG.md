# Changelog

This file lists the main changes with each version of the FyneDesk project.
More detailed release notes can be found on the [releases page](https://github.com/fyne-io/desktop/releases). 

## 0.1.2 - Ongoing

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

