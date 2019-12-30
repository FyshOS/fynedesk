<p align="center">
  <a href="https://godoc.org/fyne.io/desktop" title="GoDoc Reference" rel="nofollow"><img src="https://img.shields.io/badge/go-documentation-blue.svg?style=flat" alt="GoDoc Reference"></a>
  <a href="https://github.com/fyne-io/develop/releases/tag/v0.1.1" title="0.1.1 Release" rel="nofollow"><img src="https://img.shields.io/badge/version-0.1.1-blue.svg?style=flat" alt="0.1.1 release"></a>
  <a href='http://gophers.slack.com/messages/fyne'><img src='https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=blue' alt='Join us on Slack' /></a>
  <a href='https://fossfi.sh/support-fyneio'><img src='https://img.shields.io/badge/$-support_us-orange.svg?labelWidth=20&logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAkCAYAAADPRbkKAAAABmJLR0QA7wAyAD/CTveyAAAACXBIWXMAAA9hAAAPYQGoP6dpAAAAB3RJTUUH4wMVCQ4LeuPReAAABDFJREFUWMPVmX9oVWUYxz/3tr5jbX9oYiZWSmNgoYYlUVqraYvIamlCg8pqUBTZr38qMhZZ0A8oIqKIiLJoJtEIsR+zki1mGqQwZTYkImtlgk5Wy7jPnbv90XPgcD3n3Lvdy731wOGc+77nec73Oe/zfp/nOTeVy+WolphZC3CVH9N9OAsMAzuBLkkHkmykKu2AmQGsAjYAC4pQ+QJ4WNJg1R0ws3pgI3DTJFVPAM8AGyRNVMUBB98DLCvBzLtAh6QTwUC6gmGzuUTwAGuBR8MD6QpFzwPAypi5v4HXgRXAEg+vLQm2njSzCyoWQmbWAPwAzIqY/gm4UdJAhN5qoAuojdCbAD4BVlViBdbGgJ8A2gPwZtZsZvea2XwASd1AZ4zNNHA90FgJB9ri6FHStw7+WaAPeA3YbWaBzovAiF/nfMUC+QD4sRIOLIkZ73Hws4BHQuOneY7A2aY/CHdgL3AtsBhYD9TVVMCB0+O2h59nRJBJfej6WOj6Bj8C2VcpFoqSC/38PbArb24zQDabBWhOsFFbNAuZ2UygxZcvcHwPsEPScILecaAuYmoUaJR01MymAU8AFzm7vCxp3MxagW0xpv8EzinogLPC40A7cGpMKHQB6yX9FqE/ACyKMb8VWC0pG6E3D+gF5sbo9km6Ml0A/N3Ad8BtMeABBNwB9JrZ7Ij5PQmPuA7YZWZXmFnan1ljZrf65p2boDuQmInNrAN4I29DJUmT02C+bCtiL/QCo2Y2BIwB7wFzCuj1xGZiM7vcb9gOTAMunUTZMVPSkZCtJuBAmQngIHCupImaCPApoBVokvSrj7UDm4o0Pg84Evo97EkoVUYHXgjK6pNWwB1I5dfdZtbtjUghWSppZ0ivDvirjA4MAosCfCeFhaRcPvhQLV6QbYGhvLHZZQRvwO1hfJNJZFu9qkySNyUdyxu7pkzgc8BDknZPuSMzs8XA58AZEdPbgTZJY3ld2GABOixWXpH0YMk9sZlN96zZ6hn2F+AdYFM4IZlZLfBRQiMzGXlaUmdFmnpPSC3A814alCJjwDpJG+NuqCkC0HzgLi+L651RvgTeAu4ELnHazAJnedkwpwzvoh+4RdLPid+FMpnMDElHY8CvAd4GGiKmR4A1DvwpYHkZabLTO7KCkgYeiwG/3Iu0hoQ6fwvwh6QVwFKn2tEpgB4BPgSaJS0oFnywAnuB5yR15bHNV6HPfUnyqaSVIV0BlwEXe51zNtAYehGHgMNOyfu9F+iXND6V5UplMplXgfuAj4HPgIXAPcXsj5CcKelwNbqiVCaTOd/jrhRZJumbajiQlrQf6OZ/KkEpsQ74vQQ7w1V1QNIh4Gbg+BRs9BXi6kqsAJK+Bq52hihWxoH7/wshFDixw786vF9kmm+TtK+aDsTWQma2EOjg379/zgNO8akh73NfknSw2pv4H3Ayg0FmbTMRAAAAAElFTkSuQmCC' alt='Support Fyne.io' /></a>

  <br />
  <a href="https://goreportcard.com/report/fyne.io/desktop"><img src="https://goreportcard.com/badge/fyne.io/desktop" alt="Code Status" /></a>
  <a href="https://travis-ci.org/fyne-io/desktop"><img src="https://travis-ci.org/fyne-io/desktop.svg" alt="Build Status" /></a>
  <a href='https://coveralls.io/github/fyne-io/desktop?branch=develop'><img src='https://coveralls.io/repos/github/fyne-io/desktop/badge.svg?branch=develop' alt='Coverage Status' /></a>
</p>

# About

FyneDesk is an easy to use Linux desktop environment following material design.
It is build using the [Fyne](https://fyne.io) toolkit and is designed to be
easy to use as well as easy to develop. We use the Go language and welcome
any contributions or feedback for the project.

# Dependencies

For a full desktop experience you will also need the following external tools installed:

* xbacklight
* arandr

# Getting Started

Using standard go tools you can install Fyne's desktop using:

    go get fyne.io/desktop/cmd/fynedesk

Once installed you could set `$GOPATH/fynedesk` as your window manager (usually
using .xinitrc).
You can also run it in an embedded X window for testing using:

    DISPLAY=:0 Xephyr :1 -screen 1280x720 &
    DISPLAY=:1 fynedesk

It should look like this:

<p align="center" markdown="1">
  <img src="desktop-dark-current.png" alt="Fyne Desktop - Dark" />
</p>

(The default wallpaper is under Creative Commons by dave souza found on
[Wikipedia](https://commons.wikimedia.org/wiki/File:Loch_Fyne_from_Tighcladich.jpg).)

If you run the command when there is a window manager running, or on
an operating system that does not support window managers (Windows or
macOS) then the app will start in UI test mode.
When loaded in this way you can run all of the features except the
controlling of windows - they will load on your main desktop.

# Runner

A desktop needs to be rock solid and, whilst we are working hard to get there,
any alpha or beta software can run in to unexpected issues. 
For that reason we have included a `fynedesk_runner` utility that can help
manage unexpected events. If you start the desktop using the runner then
if a crash occurs it will normally recover where it left off with no loss
of data in your applications.

Using standard go tools you can install the runner using:

    go get fyne.io/desktop/cmd/fynedesk_runner

From then on execute that instead of the `fynedesk` command for a more 
resillient desktop when testing out pre-release builds.

