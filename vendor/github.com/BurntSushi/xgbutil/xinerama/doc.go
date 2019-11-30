/*
Package xinerama provides a convenience function to retrieve the geometry of
all active heads sorted in order from left to right and then top to bottom.
While Xinerama is an old extension that isn't often used in lieu of RandR and
TwinView, it can still be used to query for information about all active heads.
That is, even if TwinView or RandR is being used, Xinerama will still report
the correct geometry of each head.
*/
package xinerama
