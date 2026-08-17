// Package colorize holds the color-output primitives shared by every
// package that renders to a terminal: the ColorMode setting (always, never,
// auto with piped-output detection) and the switch that decides whether
// color is enabled. It is a leaf package so that both the model
// (checker/rules severity rendering) and the presentation layer
// (formatters) can use it without either owning it.
package colorize
