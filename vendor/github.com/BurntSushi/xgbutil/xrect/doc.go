/*
Package xrect defines a Rect interface and an XRect type implementing the Rect
interface for working with X rectangles. Namely, X rectangles are specified by
the 4-tuple (x, y, width, height) where the origin is the top-left corner and
the width and height *must* be non-zero.

Some of the main features of this package include finding the area of
intersection of two rectangles, finding the largest overlap between some
rectangle and a set of rectangles, applying partial struts to rectangles
representing all active heads, and a function to subtract two rectangles.
*/
package xrect
